package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/tanq16/box/internal/client"
	"github.com/tanq16/box/internal/types"
)

func PurgeFile(c *client.BoxClient, fileID string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/files/%s/trash", client.APIBaseURL, fileID), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("failed to purge file: %w", err)
	}
	defer func() { io.Copy(io.Discard, resp.Body); resp.Body.Close() }()
	if resp.StatusCode != http.StatusNoContent {
		return client.HandleError("purge file", resp)
	}
	return nil
}

func PurgeFolder(c *client.BoxClient, folderID string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/folders/%s/trash", client.APIBaseURL, folderID), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("failed to purge folder: %w", err)
	}
	defer func() { io.Copy(io.Discard, resp.Body); resp.Body.Close() }()
	if resp.StatusCode != http.StatusNoContent {
		return client.HandleError("purge folder", resp)
	}
	return nil
}

func ListTrash(c *client.BoxClient) ([]types.BoxItemDisplay, error) {
	var items []types.BoxItemDisplay
	marker := ""
	for {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/folders/trash/items", client.APIBaseURL), nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		q := req.URL.Query()
		q.Add("fields", "type,name,id,size,modified_at")
		q.Add("usemarker", "true")
		q.Add("limit", "1000")
		if marker != "" {
			q.Add("marker", marker)
		}
		req.URL.RawQuery = q.Encode()

		resp, err := c.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to list trash: %w", err)
		}
		if resp.StatusCode != http.StatusOK {
			err := client.HandleError("list trash", resp)
			resp.Body.Close()
			return nil, err
		}
		var page types.BoxFolderItems
		err = json.NewDecoder(resp.Body).Decode(&page)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		for _, item := range page.Entries {
			display := types.BoxItemDisplay{ID: item.ID, Name: item.Name, Type: item.Type}
			if item.ModifiedAt != nil {
				if t, err := time.Parse(time.RFC3339, *item.ModifiedAt); err == nil {
					display.ModifiedTime = t.Format("2006-01-02 15:04")
				}
			}
			if item.Size != nil {
				display.Size = *item.Size
			}
			items = append(items, display)
		}
		if page.NextMarker == "" {
			break
		}
		marker = page.NextMarker
	}
	return items, nil
}

func ResolveTrashItem(c *client.BoxClient, name string, itemID string) (string, string, error) {
	items, err := ListTrash(c)
	if err != nil {
		return "", "", err
	}
	return selectTrashItem(items, name, itemID)
}

func selectTrashItem(items []types.BoxItemDisplay, name string, itemID string) (string, string, error) {
	if itemID != "" {
		for _, item := range items {
			if item.ID == itemID {
				return item.ID, item.Type, nil
			}
		}
		return "", "", fmt.Errorf("item ID '%s' not found in trash", itemID)
	}
	var matches []types.BoxItemDisplay
	for _, item := range items {
		if strings.EqualFold(item.Name, name) {
			matches = append(matches, item)
		}
	}
	switch len(matches) {
	case 0:
		return "", "", fmt.Errorf("no trashed item named '%s'", name)
	case 1:
		return matches[0].ID, matches[0].Type, nil
	default:
		return "", "", fmt.Errorf("multiple trashed items named '%s'; use --id to disambiguate", name)
	}
}

func RestoreItem(c *client.BoxClient, itemType string, itemID string, parentID string) (*types.BoxItem, error) {
	endpoint := "files"
	if itemType == "folder" {
		endpoint = "folders"
	}
	bodyMap := map[string]any{}
	if parentID != "" {
		bodyMap["parent"] = map[string]string{"id": parentID}
	}
	bodyBytes, err := json.Marshal(bodyMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s/%s", client.APIBaseURL, endpoint, itemID), bytes.NewBuffer(bodyBytes))
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

func EmptyTrash(ctx context.Context, c *client.BoxClient) (int, error) {
	items, err := ListTrash(c)
	if err != nil {
		return 0, err
	}
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(8)
	var purged atomic.Int64
	for _, item := range items {
		g.Go(func() error {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			var err error
			if item.Type == "folder" {
				err = PurgeFolder(c, item.ID)
			} else {
				err = PurgeFile(c, item.ID)
			}
			if err != nil {
				return fmt.Errorf("failed to purge %s '%s': %w", item.Type, item.Name, err)
			}
			purged.Add(1)
			return nil
		})
	}
	err = g.Wait()
	return int(purged.Load()), err
}
