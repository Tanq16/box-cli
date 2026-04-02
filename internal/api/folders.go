package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/tanq16/box/internal/client"
	"github.com/tanq16/box/internal/types"
)

func ResolvePath(c *client.BoxClient, path string, expectedType string) (string, string, error) {
	if path == "" || path == "/" || path == "root" {
		return "0", "folder", nil
	}
	segments := strings.Split(strings.Trim(path, "/"), "/")
	currentID := "0"
	var currentType string

	for i, segment := range segments {
		if segment == "" {
			continue
		}
		isLastSegment := (i == len(segments) - 1)

		found := false
		offset := 0
		limit := 1000

		for !found {
			req, err := http.NewRequest("GET", fmt.Sprintf("%s/folders/%s/items", client.APIBaseURL, currentID), nil)
			if err != nil {
				return "", "", fmt.Errorf("failed to create request: %w", err)
			}
			q := req.URL.Query()
			q.Add("fields", "type,name,id")
			q.Add("limit", fmt.Sprintf("%d", limit))
			q.Add("offset", fmt.Sprintf("%d", offset))
			req.URL.RawQuery = q.Encode()

			resp, err := c.Do(req)
			if err != nil {
				return "", "", fmt.Errorf("failed to list folder %s: %w", currentID, err)
			}
			if resp.StatusCode != http.StatusOK {
				err := client.HandleError("resolve path", resp)
				resp.Body.Close()
				return "", "", err
			}
			var items types.BoxFolderItems
			err = json.NewDecoder(resp.Body).Decode(&items)
			resp.Body.Close()
			if err != nil {
				return "", "", fmt.Errorf("failed to parse response: %w", err)
			}

			for _, item := range items.Entries {
				if strings.EqualFold(item.Name, segment) {
					currentID = item.ID
					currentType = item.Type
					found = true
					break
				}
			}
			if found {
				break
			}
			offset += len(items.Entries)
			if offset >= items.TotalCount || len(items.Entries) == 0 {
				break
			}
		}

		if !found {
			return "", "", fmt.Errorf("path not found: '%s' in '%s'", segment, path)
		}
		if isLastSegment {
			if expectedType != "" && currentType != expectedType {
				return "", "", fmt.Errorf("'%s' is a %s, but expected a %s", segment, currentType, expectedType)
			}
			return currentID, currentType, nil
		}
		if currentType != "folder" {
			return "", "", fmt.Errorf("'%s' in '%s' is a file, not a folder", segment, path)
		}
	}
	return "0", "folder", nil
}

func ListFolder(c *client.BoxClient, folderID string) ([]types.BoxItemDisplay, []types.BoxItemDisplay, error) {
	var allFolders []types.BoxItemDisplay
	var allFiles []types.BoxItemDisplay
	offset := 0
	limit := 1000

	for {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/folders/%s/items", client.APIBaseURL, folderID), nil)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create request: %w", err)
		}
		q := req.URL.Query()
		q.Add("fields", "type,name,id,size,modified_at")
		q.Add("limit", fmt.Sprintf("%d", limit))
		q.Add("offset", fmt.Sprintf("%d", offset))
		req.URL.RawQuery = q.Encode()

		resp, err := c.Do(req)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to list folder: %w", err)
		}
		if resp.StatusCode != http.StatusOK {
			err := client.HandleError("list folder", resp)
			resp.Body.Close()
			return nil, nil, err
		}
		var items types.BoxFolderItems
		err = json.NewDecoder(resp.Body).Decode(&items)
		resp.Body.Close()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse response: %w", err)
		}

		for _, item := range items.Entries {
			var modTime string
			if item.ModifiedAt != nil {
				t, err := time.Parse(time.RFC3339, *item.ModifiedAt)
				if err == nil {
					modTime = t.Format("2006-01-02 15:04")
				}
			}
			display := types.BoxItemDisplay{
				ID: item.ID, Name: item.Name, ModifiedTime: modTime, Type: item.Type,
			}
			if item.Size != nil {
				display.Size = *item.Size
			}
			if item.Type == "folder" {
				allFolders = append(allFolders, display)
			} else {
				allFiles = append(allFiles, display)
			}
		}
		offset += len(items.Entries)
		if offset >= items.TotalCount || len(items.Entries) == 0 {
			break
		}
	}
	sort.Slice(allFolders, func(i, j int) bool { return allFolders[i].Name < allFolders[j].Name })
	sort.Slice(allFiles, func(i, j int) bool { return allFiles[i].Name < allFiles[j].Name })
	return allFolders, allFiles, nil
}

func FindOrCreateFolder(c *client.BoxClient, folderName string, parentID string) (string, error) {
	offset := 0
	limit := 1000
	for {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/folders/%s/items", client.APIBaseURL, parentID), nil)
		if err != nil {
			return "", fmt.Errorf("failed to create request: %w", err)
		}
		q := req.URL.Query()
		q.Add("fields", "type,name,id")
		q.Add("limit", fmt.Sprintf("%d", limit))
		q.Add("offset", fmt.Sprintf("%d", offset))
		req.URL.RawQuery = q.Encode()

		resp, err := c.Do(req)
		if err != nil {
			return "", fmt.Errorf("failed to list folder: %w", err)
		}
		if resp.StatusCode != http.StatusOK {
			err := client.HandleError("find folder", resp)
			resp.Body.Close()
			return "", err
		}
		var items types.BoxFolderItems
		if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
			resp.Body.Close()
			return "", fmt.Errorf("failed to parse response: %w", err)
		}
		resp.Body.Close()

		for _, item := range items.Entries {
			if strings.EqualFold(item.Name, folderName) && item.Type == "folder" {
				return item.ID, nil
			}
		}
		offset += len(items.Entries)
		if offset >= items.TotalCount || len(items.Entries) == 0 {
			break
		}
	}

	body := fmt.Sprintf(`{"name":%q,"parent":{"id":%q}}`, folderName, parentID)
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/folders", client.APIBaseURL), bytes.NewBufferString(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to create folder: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return "", client.HandleError("create folder", resp)
	}
	var folder types.BoxItem
	if err := json.NewDecoder(resp.Body).Decode(&folder); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}
	return folder.ID, nil
}

func GetFolderInfo(c *client.BoxClient, folderID string) (*types.BoxItem, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/folders/%s", client.APIBaseURL, folderID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	q := req.URL.Query()
	q.Add("fields", "type,name,id,size,modified_at,parent,item_collection")
	req.URL.RawQuery = q.Encode()

	var item types.BoxItem
	_, err = c.DoJSON(req, &item)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func CreateFolder(c *client.BoxClient, folderName string, parentID string) (*types.BoxItem, error) {
	body := fmt.Sprintf(`{"name":%q,"parent":{"id":%q}}`, folderName, parentID)
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/folders", client.APIBaseURL), bytes.NewBufferString(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create folder: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return nil, client.HandleError("create folder", resp)
	}
	var folder types.BoxItem
	if err := json.NewDecoder(resp.Body).Decode(&folder); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &folder, nil
}

func MoveItem(c *client.BoxClient, itemType string, itemID string, newParentID string) (*types.BoxItem, error) {
	endpoint := "files"
	if itemType == "folder" {
		endpoint = "folders"
	}
	body := fmt.Sprintf(`{"parent":{"id":%q}}`, newParentID)
	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/%s/%s", client.APIBaseURL, endpoint, itemID), bytes.NewBufferString(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	var item types.BoxItem
	_, err = c.DoJSON(req, &item)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func RenameItem(c *client.BoxClient, itemType string, itemID string, newName string) (*types.BoxItem, error) {
	endpoint := "files"
	if itemType == "folder" {
		endpoint = "folders"
	}
	body := fmt.Sprintf(`{"name":%q}`, newName)
	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/%s/%s", client.APIBaseURL, endpoint, itemID), bytes.NewBufferString(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	var item types.BoxItem
	_, err = c.DoJSON(req, &item)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func CopyFile(c *client.BoxClient, fileID string, destFolderID string, newName string) (*types.BoxItem, error) {
	bodyMap := map[string]interface{}{
		"parent": map[string]string{"id": destFolderID},
	}
	if newName != "" {
		bodyMap["name"] = newName
	}
	bodyBytes, err := json.Marshal(bodyMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/files/%s/copy", client.APIBaseURL, fileID), bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	var item types.BoxItem
	_, err = c.DoJSON(req, &item)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func CopyFolder(c *client.BoxClient, folderID string, destFolderID string, newName string) (*types.BoxItem, error) {
	bodyMap := map[string]interface{}{
		"parent": map[string]string{"id": destFolderID},
	}
	if newName != "" {
		bodyMap["name"] = newName
	}
	bodyBytes, err := json.Marshal(bodyMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/folders/%s/copy", client.APIBaseURL, folderID), bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	var item types.BoxItem
	_, err = c.DoJSON(req, &item)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func DeleteFolder(c *client.BoxClient, folderID string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/folders/%s", client.APIBaseURL, folderID), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	q := req.URL.Query()
	q.Add("recursive", "true")
	req.URL.RawQuery = q.Encode()

	resp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete folder: %w", err)
	}
	defer func() { io.Copy(io.Discard, resp.Body); resp.Body.Close() }()
	if resp.StatusCode != http.StatusNoContent {
		return client.HandleError("delete folder", resp)
	}
	return nil
}

func DeleteFile(c *client.BoxClient, fileID string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/files/%s", client.APIBaseURL, fileID), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	defer func() { io.Copy(io.Discard, resp.Body); resp.Body.Close() }()
	if resp.StatusCode != http.StatusNoContent {
		return client.HandleError("delete file", resp)
	}
	return nil
}
