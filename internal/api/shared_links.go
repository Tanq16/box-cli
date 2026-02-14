package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/tanq16/box/internal/client"
	"github.com/tanq16/box/internal/types"
)

// CreateSharedLink creates a shared link on a file or folder.
func CreateSharedLink(c *client.BoxClient, itemType, itemID, access, password string) (*types.BoxItem, error) {
	endpoint := "files"
	if itemType == "folder" {
		endpoint = "folders"
	}

	sharedLink := map[string]interface{}{
		"access": access,
	}
	if password != "" {
		sharedLink["password"] = password
	}

	payload := map[string]interface{}{
		"shared_link": sharedLink,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/%s/%s", client.APIBaseURL, endpoint, itemID), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	q := req.URL.Query()
	q.Add("fields", "shared_link,name,id,type")
	req.URL.RawQuery = q.Encode()

	var item types.BoxItem
	_, err = c.DoJSON(req, &item)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// GetSharedLink retrieves the shared link for a file or folder.
func GetSharedLink(c *client.BoxClient, itemType, itemID string) (*types.BoxItem, error) {
	endpoint := "files"
	if itemType == "folder" {
		endpoint = "folders"
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s/%s", client.APIBaseURL, endpoint, itemID), nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Add("fields", "shared_link,name,id,type")
	req.URL.RawQuery = q.Encode()

	var item types.BoxItem
	_, err = c.DoJSON(req, &item)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// RemoveSharedLink removes the shared link from a file or folder.
func RemoveSharedLink(c *client.BoxClient, itemType, itemID string) error {
	endpoint := "files"
	if itemType == "folder" {
		endpoint = "folders"
	}

	body := []byte(`{"shared_link":null}`)
	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/%s/%s", client.APIBaseURL, endpoint, itemID), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	q := req.URL.Query()
	q.Add("fields", "shared_link")
	req.URL.RawQuery = q.Encode()

	var item types.BoxItem
	_, err = c.DoJSON(req, &item)
	return err
}

// ResolveSharedLink resolves a shared link URL to get item info.
func ResolveSharedLink(c *client.BoxClient, sharedURL string, password string) (*types.BoxItem, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/shared_items", client.APIBaseURL), nil)
	if err != nil {
		return nil, err
	}

	boxapiHeader := fmt.Sprintf("shared_link=%s", sharedURL)
	if password != "" {
		boxapiHeader += fmt.Sprintf("&shared_link_password=%s", password)
	}
	req.Header.Set("BoxApi", boxapiHeader)

	q := req.URL.Query()
	q.Add("fields", "type,id,name,size,modified_at,parent")
	req.URL.RawQuery = q.Encode()

	var item types.BoxItem
	_, err = c.DoJSON(req, &item)
	if err != nil {
		return nil, err
	}
	return &item, nil
}
