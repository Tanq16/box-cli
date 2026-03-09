package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/tanq16/box/internal/client"
	"github.com/tanq16/box/internal/types"
)

func CreateCollaboration(c *client.BoxClient, itemType, itemID, role, userEmail, userID, groupID string) (*types.Collaboration, error) {
	payload := map[string]interface{}{
		"item": map[string]string{
			"type": itemType,
			"id":   itemID,
		},
		"role": role,
	}

	accessibleBy := map[string]string{}
	if groupID != "" {
		accessibleBy["type"] = "group"
		accessibleBy["id"] = groupID
	} else if userID != "" {
		accessibleBy["type"] = "user"
		accessibleBy["id"] = userID
	} else if userEmail != "" {
		accessibleBy["type"] = "user"
		accessibleBy["login"] = userEmail
	} else {
		return nil, fmt.Errorf("must specify --user-email, --user-id, or --group-id")
	}
	payload["accessible_by"] = accessibleBy

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/collaborations", client.APIBaseURL), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	var collab types.Collaboration
	_, err = c.DoJSON(req, &collab)
	if err != nil {
		return nil, err
	}
	return &collab, nil
}

func GetCollaboration(c *client.BoxClient, collabID string) (*types.Collaboration, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/collaborations/%s", client.APIBaseURL, collabID), nil)
	if err != nil {
		return nil, err
	}

	var collab types.Collaboration
	_, err = c.DoJSON(req, &collab)
	if err != nil {
		return nil, err
	}
	return &collab, nil
}

func UpdateCollaboration(c *client.BoxClient, collabID, role, status string) (*types.Collaboration, error) {
	payload := map[string]string{}
	if role != "" {
		payload["role"] = role
	}
	if status != "" {
		payload["status"] = status
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/collaborations/%s", client.APIBaseURL, collabID), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	var collab types.Collaboration
	_, err = c.DoJSON(req, &collab)
	if err != nil {
		return nil, err
	}
	return &collab, nil
}

func DeleteCollaboration(c *client.BoxClient, collabID string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/collaborations/%s", client.APIBaseURL, collabID), nil)
	if err != nil {
		return err
	}

	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer func() { io.Copy(io.Discard, resp.Body); resp.Body.Close() }()
	if resp.StatusCode != http.StatusNoContent {
		return client.HandleError("delete collaboration", resp)
	}
	return nil
}

func ListPendingCollaborations(c *client.BoxClient) (*types.CollaborationList, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/collaborations", client.APIBaseURL), nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Add("status", "pending")
	req.URL.RawQuery = q.Encode()

	var list types.CollaborationList
	_, err = c.DoJSON(req, &list)
	if err != nil {
		return nil, err
	}
	return &list, nil
}
