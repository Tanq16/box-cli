package api

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/tanq16/box/internal/client"
	"github.com/tanq16/box/internal/types"
)

func UploadFileChunked(c *client.BoxClient, localPath string, parentFolderID string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}
	fileSize := stat.Size()
	fileName := filepath.Base(localPath)

	session, err := createUploadSession(c, fileName, parentFolderID, fileSize)
	if err != nil {
		return fmt.Errorf("failed to create upload session: %w", err)
	}

	log.Debug().Int("total_parts", session.TotalParts).Int64("part_size", session.PartSize).Msg("upload session created")

	file.Seek(0, io.SeekStart)
	wholeFileHash := sha1.New()

	var parts []types.UploadPart
	partSize := session.PartSize
	offset := int64(0)

	for partNum := 0; offset < fileSize; partNum++ {
		currentPartSize := partSize
		if offset+currentPartSize > fileSize {
			currentPartSize = fileSize - offset
		}

		buf := make([]byte, currentPartSize)
		n, err := io.ReadFull(file, buf)
		if err != nil && err != io.ErrUnexpectedEOF {
			abortUploadSession(c, session.ID)
			return fmt.Errorf("failed to read part %d: %w", partNum, err)
		}
		buf = buf[:n]

		wholeFileHash.Write(buf)

		part, err := uploadPart(c, session.ID, buf, offset, fileSize)
		if err != nil {
			abortUploadSession(c, session.ID)
			return fmt.Errorf("failed to upload part %d: %w", partNum, err)
		}
		parts = append(parts, *part)
		offset += int64(n)

		log.Debug().Int("part", partNum+1).Int("total", session.TotalParts).Msg("uploaded part")
	}

	wholeDigest := base64.StdEncoding.EncodeToString(wholeFileHash.Sum(nil))
	if err := commitUploadSession(c, session.ID, parts, wholeDigest); err != nil {
		abortUploadSession(c, session.ID)
		return fmt.Errorf("failed to commit upload: %w", err)
	}

	log.Debug().Msg("chunked upload complete")
	return nil
}

func createUploadSession(c *client.BoxClient, fileName string, folderID string, fileSize int64) (*types.UploadSession, error) {
	body := fmt.Sprintf(`{"folder_id":%q,"file_size":%d,"file_name":%q}`, folderID, fileSize, fileName)
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/files/upload_sessions", client.UploadBaseURL), bytes.NewBufferString(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		var conflictErr struct {
			ContextInfo struct {
				Conflicts []types.BoxItem `json:"conflicts"`
			} `json:"context_info"`
		}
		respBody, _ := io.ReadAll(resp.Body)
		if json.Unmarshal(respBody, &conflictErr) == nil && len(conflictErr.ContextInfo.Conflicts) > 0 {
			existingID := conflictErr.ContextInfo.Conflicts[0].ID
			log.Debug().Str("existing_id", existingID).Msg("file exists, creating version upload session")
			return createVersionUploadSession(c, existingID, fileSize)
		}
		return nil, fmt.Errorf("upload session conflict for '%s'", fileName)
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, client.HandleError("create upload session", resp)
	}

	var session types.UploadSession
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return nil, fmt.Errorf("failed to decode upload session: %w", err)
	}
	return &session, nil
}

func createVersionUploadSession(c *client.BoxClient, fileID string, fileSize int64) (*types.UploadSession, error) {
	body := fmt.Sprintf(`{"file_size":%d}`, fileSize)
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/files/%s/upload_sessions", client.UploadBaseURL, fileID), bytes.NewBufferString(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	var session types.UploadSession
	_, err = c.DoJSON(req, &session)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func uploadPart(c *client.BoxClient, sessionID string, data []byte, offset int64, totalSize int64) (*types.UploadPart, error) {
	h := sha1.New()
	h.Write(data)
	digest := "sha=" + base64.StdEncoding.EncodeToString(h.Sum(nil))

	end := offset + int64(len(data)) - 1
	contentRange := fmt.Sprintf("bytes %d-%d/%d", offset, end, totalSize)

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/files/upload_sessions/%s", client.UploadBaseURL, sessionID), bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Digest", digest)
	req.Header.Set("Content-Range", contentRange)

	var partResp types.UploadPartResponse
	_, err = c.DoJSON(req, &partResp)
	if err != nil {
		return nil, err
	}
	return &partResp.Part, nil
}

func commitUploadSession(c *client.BoxClient, sessionID string, parts []types.UploadPart, wholeFileDigest string) error {
	payload := struct {
		Parts []types.UploadPart `json:"parts"`
	}{Parts: parts}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	for attempt := 0; attempt < 5; attempt++ {
		req, err := http.NewRequest("POST", fmt.Sprintf("%s/files/upload_sessions/%s/commit", client.UploadBaseURL, sessionID), bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Digest", "sha="+wholeFileDigest)

		resp, err := c.Do(req)
		if err != nil {
			return err
		}
		resp.Body.Close()

		if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
			return nil
		}
		if resp.StatusCode == http.StatusAccepted {
			wait := 2
			if ra := resp.Header.Get("Retry-After"); ra != "" {
				if v, err := strconv.Atoi(ra); err == nil {
					wait = v
				}
			}
			log.Debug().Int("wait_seconds", wait).Msg("commit pending, retrying")
			time.Sleep(time.Duration(wait) * time.Second)
			continue
		}
		return fmt.Errorf("commit failed with status %d", resp.StatusCode)
	}
	return fmt.Errorf("commit timed out after retries")
}

func abortUploadSession(c *client.BoxClient, sessionID string) {
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/files/upload_sessions/%s", client.UploadBaseURL, sessionID), nil)
	resp, err := c.Do(req)
	if err == nil {
		resp.Body.Close()
	}
}
