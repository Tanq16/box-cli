package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/tanq16/box/internal/client"
	"github.com/tanq16/box/internal/types"
)

// GetFileInfo retrieves file metadata by ID.
func GetFileInfo(c *client.BoxClient, fileID string) (*types.BoxItem, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/files/%s", client.APIBaseURL, fileID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	q := req.URL.Query()
	q.Add("fields", "type,name,id,size,sha1,modified_at,parent")
	req.URL.RawQuery = q.Encode()

	var item types.BoxItem
	_, err = c.DoJSON(req, &item)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// UploadFile uploads a single file to a Box folder using multipart POST.
func UploadFile(c *client.BoxClient, localPath string, parentFolderID string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open file '%s': %w", localPath, err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fileName := filepath.Base(localPath)

	attributesJSON := fmt.Sprintf(`{"name":%q,"parent":{"id":%q}}`, fileName, parentFolderID)
	if err := writer.WriteField("attributes", attributesJSON); err != nil {
		return fmt.Errorf("failed to write attributes: %w", err)
	}

	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("failed to copy file data: %w", err)
	}
	writer.Close()

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/files/content", client.UploadBaseURL), body)
	if err != nil {
		return fmt.Errorf("failed to create upload request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		// File already exists — upload new version
		var conflictErr struct {
			ContextInfo struct {
				Conflicts []types.BoxItem `json:"conflicts"`
			} `json:"context_info"`
		}
		respBody, _ := io.ReadAll(resp.Body)
		if json.Unmarshal(respBody, &conflictErr) == nil && len(conflictErr.ContextInfo.Conflicts) > 0 {
			existingID := conflictErr.ContextInfo.Conflicts[0].ID
			return UploadFileVersion(c, localPath, existingID)
		}
		return fmt.Errorf("upload conflict for '%s'", fileName)
	}

	if resp.StatusCode != http.StatusCreated {
		return client.HandleError("upload file", resp)
	}
	return nil
}

// UploadFileVersion uploads a new version of an existing file.
func UploadFileVersion(c *client.BoxClient, localPath string, fileID string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open file '%s': %w", localPath, err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fileName := filepath.Base(localPath)

	attributesJSON := fmt.Sprintf(`{"name":%q}`, fileName)
	if err := writer.WriteField("attributes", attributesJSON); err != nil {
		return fmt.Errorf("failed to write attributes: %w", err)
	}

	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("failed to copy file data: %w", err)
	}
	writer.Close()

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/files/%s/content", client.UploadBaseURL, fileID), body)
	if err != nil {
		return fmt.Errorf("failed to create upload request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload file version: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return client.HandleError("upload file version", resp)
	}
	return nil
}

// UploadFolder recursively uploads a local directory to Box.
func UploadFolder(c *client.BoxClient, localPath string, parentFolderID string) error {
	rootFolderName := filepath.Base(localPath)
	rootBoxID, err := FindOrCreateFolder(c, rootFolderName, parentFolderID)
	if err != nil {
		return fmt.Errorf("failed to create root folder '%s': %w", rootFolderName, err)
	}

	folderIDMap := make(map[string]string)
	folderIDMap[localPath] = rootBoxID

	return filepath.WalkDir(localPath, func(currentPath string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if currentPath == localPath {
			return nil
		}

		parentDir := filepath.Dir(currentPath)
		parentBoxID, ok := folderIDMap[parentDir]
		if !ok {
			return fmt.Errorf("could not find parent Box ID for: %s", parentDir)
		}

		if d.IsDir() {
			boxFolderID, err := FindOrCreateFolder(c, d.Name(), parentBoxID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to create folder '%s': %v\n", d.Name(), err)
				return filepath.SkipDir
			}
			folderIDMap[currentPath] = boxFolderID
		} else {
			if err := UploadFile(c, currentPath, parentBoxID); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to upload '%s': %v\n", currentPath, err)
			} else {
				fmt.Fprintf(os.Stderr, "Uploaded: %s\n", currentPath)
			}
		}
		return nil
	})
}

// DownloadFile downloads a file by ID to a local path.
func DownloadFile(c *client.BoxClient, fileID string, localPath string) (string, error) {
	// If no local path, get filename from API
	if localPath == "" {
		info, err := GetFileInfo(c, fileID)
		if err == nil && info.Name != "" {
			localPath = info.Name
		} else {
			localPath = "downloaded_file"
		}
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/files/%s/content", client.APIBaseURL, fileID), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create download request: %w", err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", client.HandleError("download file", resp)
	}

	out, err := os.Create(localPath)
	if err != nil {
		return "", fmt.Errorf("failed to create local file '%s': %w", localPath, err)
	}
	defer out.Close()
	if _, err := io.Copy(out, resp.Body); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}
	return localPath, nil
}

// DownloadFolder recursively downloads a folder to a local directory.
func DownloadFolder(c *client.BoxClient, folderID string, localPath string) error {
	if err := os.MkdirAll(localPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory '%s': %w", localPath, err)
	}

	folders, files, err := ListFolder(c, folderID)
	if err != nil {
		return fmt.Errorf("failed to list folder: %w", err)
	}

	for _, file := range files {
		destPath := filepath.Join(localPath, file.Name)
		_, err := DownloadFile(c, file.ID, destPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to download '%s': %v\n", file.Name, err)
		} else {
			fmt.Fprintf(os.Stderr, "Downloaded: %s\n", destPath)
		}
	}

	for _, folder := range folders {
		destPath := filepath.Join(localPath, folder.Name)
		if err := DownloadFolder(c, folder.ID, destPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to download folder '%s': %v\n", folder.Name, err)
		}
	}
	return nil
}

// ResolveRemoteFileID resolves a remote path or uses the --id flag to get a file ID.
func ResolveRemoteFileID(c *client.BoxClient, remotePath string, directID string) (string, error) {
	if directID != "" {
		return directID, nil
	}
	if remotePath == "" {
		return "", fmt.Errorf("must specify a remote path or --id")
	}
	id, itemType, err := ResolvePath(c, remotePath, "")
	if err != nil {
		return "", err
	}
	if itemType != "file" {
		return "", fmt.Errorf("'%s' is a %s, not a file", remotePath, itemType)
	}
	return id, nil
}

// ResolveRemoteFolderID resolves a remote path or uses the --id flag to get a folder ID.
func ResolveRemoteFolderID(c *client.BoxClient, remotePath string, directID string) (string, error) {
	if directID != "" {
		return directID, nil
	}
	remotePath = strings.TrimRight(remotePath, "/")
	if remotePath == "" {
		return "0", nil
	}
	id, _, err := ResolvePath(c, remotePath, "folder")
	if err != nil {
		return "", err
	}
	return id, nil
}
