# Reference Plan: Go-Based Box CLI Tool

> A standalone Go binary for Box operations, built on direct REST API calls.
> Cross-compilable, single-binary, no Node.js/NPM dependency.

---

## Table of Contents

1. [Architecture Overview](#1-architecture-overview)
2. [Project Structure](#2-project-structure)
3. [Go Dependencies](#3-go-dependencies)
4. [Reference Code from ANBU — OAuth & Auth](#4-reference-code-from-anbu--oauth--auth)
5. [Reference Code from ANBU — Primitives & Constants](#5-reference-code-from-anbu--primitives--constants)
6. [Reference Code from ANBU — File & Folder Operations](#6-reference-code-from-anbu--file--folder-operations)
7. [Reference Code from ANBU — Sync with SHA1 Comparison](#7-reference-code-from-anbu--sync-with-sha1-comparison)
8. [Reference Code from ANBU — Local Indexing & Search](#8-reference-code-from-anbu--local-indexing--search)
9. [Reference Code from ANBU — CLI Command Definitions](#9-reference-code-from-anbu--cli-command-definitions)
10. [New Operations — API Specs](#10-new-operations--api-specs)
    - [Collaborations](#collaborations)
    - [Comments](#comments)
    - [Shared Links](#shared-links)
    - [Trash Operations](#trash-operations)
    - [Search (Server-Side)](#search-server-side)
    - [Users Management](#users-management)
    - [Groups Management](#groups-management)
    - [Chunked Uploads](#chunked-uploads)
11. [Common Error Response Schema](#11-common-error-response-schema)
12. [Option C — OpenAPI Generated Go Client](#12-option-c--openapi-generated-go-client)
13. [Build & Cross-Compilation](#13-build--cross-compilation)

---

## 1. Architecture Overview

```
┌─────────────────────────────────────────────┐
│            CLI Layer (cobra)                 │
│  Commands: list, upload, download, sync,    │
│  search, collab, comment, trash, users, ... │
├─────────────────────────────────────────────┤
│         Box API Client Layer                │
│  Direct REST calls via net/http             │
│  OR generated client from OpenAPI spec      │
├─────────────────────────────────────────────┤
│         Auth Layer (golang.org/x/oauth2)    │
│  OAuth 2.0 with token caching & refresh     │
├─────────────────────────────────────────────┤
│         Box REST API                        │
│  https://api.box.com/2.0/                   │
│  https://upload.box.com/api/2.0/            │
└─────────────────────────────────────────────┘
```

**Key insight:** The Box SDK (Node.js) is purely a wrapper around the REST API. There is NO proprietary functionality. Every SDK operation maps to documented REST endpoints. The OpenAPI 3.0 spec at `github.com/box/box-openapi` is the source of truth and covers 100% of the public API surface. There is no official Go SDK.

---

## 2. Project Structure

```
boxctl/
├── main.go
├── go.mod
├── go.sum
├── cmd/
│   ├── root.go              # Root cobra command
│   ├── login.go             # OAuth login command
│   ├── list.go              # List files/folders
│   ├── upload.go            # Upload file/folder
│   ├── download.go          # Download file/folder
│   ├── sync.go              # Sync directory
│   ├── search.go            # Server-side search
│   ├── index.go             # Local indexing
│   ├── collab.go            # Collaborations CRUD
│   ├── comment.go           # Comments CRUD
│   ├── shared_link.go       # Shared link operations
│   ├── trash.go             # Trash operations
│   ├── users.go             # User management
│   └── groups.go            # Group management
├── internal/
│   ├── auth/
│   │   └── oauth.go         # OAuth flow, token cache, refresh
│   ├── client/
│   │   └── client.go        # HTTP client wrapper, retry, error handling
│   ├── api/
│   │   ├── files.go         # File operations
│   │   ├── folders.go       # Folder operations
│   │   ├── upload.go        # Upload (simple + chunked)
│   │   ├── download.go      # Download operations
│   │   ├── sync.go          # Sync logic
│   │   ├── search.go        # Server-side search
│   │   ├── index.go         # Local indexing
│   │   ├── collab.go        # Collaboration operations
│   │   ├── comments.go      # Comment operations
│   │   ├── shared_links.go  # Shared link operations
│   │   ├── trash.go         # Trash operations
│   │   ├── users.go         # User operations
│   │   └── groups.go        # Group operations
│   └── types/
│       └── types.go         # All struct definitions
└── Makefile                  # Cross-compilation targets
```

---

## 3. Go Dependencies

```go
module github.com/tanq16/boxctl

go 1.23

require (
    github.com/spf13/cobra v1.8.0    // CLI framework
    golang.org/x/oauth2 v0.27.0      // OAuth 2.0
    github.com/rs/zerolog v1.32.0     // Structured logging (optional)
)
```

---

## 4. Reference Code from ANBU — OAuth & Auth

Source: `anbu/internal/interactions/box/auth.go`

```go
package box

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"net/http"

	"github.com/rs/zerolog/log"
	u "github.com/tanq16/anbu/utils"
	"golang.org/x/oauth2"
)

func GetBoxClient(credentialsFile string) (*http.Client, error) {
	ctx := context.Background()
	b, err := os.ReadFile(credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read credentials file: %v", err)
	}
	var creds BoxCredentials
	if err := json.Unmarshal(b, &creds); err != nil {
		return nil, fmt.Errorf("unable to parse credentials file: %v", err)
	}
	if creds.ClientID == "" || creds.ClientSecret == "" {
		return nil, fmt.Errorf("credentials file must contain client_id and client_secret")
	}
	config := &oauth2.Config{
		ClientID:     creds.ClientID,
		ClientSecret: creds.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://account.box.com/api/oauth2/authorize",
			TokenURL: "https://api.box.com/oauth2/token",
		},
		RedirectURL: redirectURI,
		Scopes:      []string{"root_readwrite"},
	}
	token, err := getBoxOAuthToken(config)
	if err != nil {
		return nil, fmt.Errorf("unable to get OAuth token: %v", err)
	}
	tokenSource := config.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("unable to refresh token: %v", err)
	}
	if newToken.AccessToken != token.AccessToken {
		log.Debug().Str("op", "box/auth").Msg("access token was refreshed")
		saveBoxToken(newToken)
	}
	return oauth2.NewClient(ctx, tokenSource), nil
}

func getBoxOAuthToken(config *oauth2.Config) (*oauth2.Token, error) {
	tokenFile, err := getBoxTokenFilePath()
	if err != nil {
		return nil, err
	}
	token, err := boxTokenFromFile(tokenFile)
	if err == nil {
		if token.Valid() {
			log.Debug().Str("op", "box/auth").Msg("existing token retrieved and valid")
			return token, nil
		}
		if token.RefreshToken != "" {
			log.Debug().Str("op", "box/auth").Msg("refreshing expired token")
			tokenSource := config.TokenSource(context.Background(), token)
			newToken, err := tokenSource.Token()
			if err != nil {
				return nil, fmt.Errorf("unable to refresh token: %v", err)
			}
			token = newToken
			if err := saveBoxToken(token); err != nil {
				// warning: unable to save refreshed token
			}
			return token, nil
		}
	}
	log.Debug().Str("op", "box/auth").Msg("no valid token, starting new OAuth flow")
	state := fmt.Sprintf("st%d", os.Getpid())
	authURL := config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	// DeviceCodeFlow: prints URL, user authorizes in browser, pastes redirect URL back
	redirectURLStr := u.DeviceCodeFlow(authURL, "")
	parsedURL, err := url.Parse(redirectURLStr)
	if err != nil {
		return nil, fmt.Errorf("could not parse the pasted URL: %v", err)
	}
	code := parsedURL.Query().Get("code")
	returnedState := parsedURL.Query().Get("state")
	if code == "" {
		return nil, fmt.Errorf("pasted URL did not contain an authorization 'code'")
	}
	if returnedState != state {
		return nil, fmt.Errorf("CSRF state mismatch. Expected '%s' but got '%s'", state, returnedState)
	}
	token, err = config.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("unable to exchange auth code for token: %v", err)
	}
	if err := saveBoxToken(token); err != nil {
		// warning: unable to save new token
	}
	return token, nil
}

func getBoxTokenFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	anbuDir := filepath.Join(homeDir, ".anbu")
	if err := os.MkdirAll(anbuDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create .anbu directory: %w", err)
	}
	return filepath.Join(anbuDir, boxTokenFile), nil
}

func boxTokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	token := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(token)
	return token, err
}

func saveBoxToken(token *oauth2.Token) error {
	tokenFile, err := getBoxTokenFilePath()
	if err != nil {
		return err
	}
	f, err := os.OpenFile(tokenFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("unable to cache oauth token: %v", err)
	}
	defer f.Close()
	err = json.NewEncoder(f).Encode(token)
	if err != nil {
		return fmt.Errorf("unable to encode token: %v", err)
	}
	return nil
}
```

### Auth Endpoints Reference

| Step | Method | URL |
|------|--------|-----|
| Authorization | Browser redirect | `https://account.box.com/api/oauth2/authorize` |
| Token Exchange | POST | `https://api.box.com/oauth2/token` |
| Redirect URI | — | `http://localhost:8080` |
| Scope | — | `root_readwrite` |

### Token Storage

- **Credentials:** `~/.boxctl/credentials.json` (or `~/.anbu/box-credentials.json`)
- **Token cache:** `~/.boxctl/token.json` (permissions `0600`)

---

## 5. Reference Code from ANBU — Primitives & Constants

Source: `anbu/internal/interactions/box/primitives.go`

```go
package box

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
)

const (
	boxTokenFile         = "box-token.json"
	redirectURI          = "http://localhost:8080"
	apiBaseURL           = "https://api.box.com/2.0"
	uploadBaseURL        = "https://upload.box.com/api/2.0"
	folderItemsURL       = apiBaseURL + "/folders/%s/items"
	fileContentURL       = apiBaseURL + "/files/%s/content"
	uploadFileURL        = uploadBaseURL + "/files/content"
	uploadFileVersionURL = uploadBaseURL + "/files/%s/content"
	uploadFolderURL      = apiBaseURL + "/folders"
)

type BoxCredentials struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type BoxItem struct {
	Type   string `json:"type"`
	ID     string `json:"id"`
	Name   string `json:"name"`
	Size   *int64 `json:"size"`
	Parent *struct {
		ID string `json:"id"`
	} `json:"parent"`
	ModifiedAt *string `json:"modified_at"`
}

type BoxFolderItems struct {
	TotalCount int       `json:"total_count"`
	Entries    []BoxItem `json:"entries"`
}

type BoxError struct {
	Type    string `json:"type"`
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type BoxItemDisplay struct {
	ID           string
	Name         string
	ModifiedTime string
	Size         int64
	Type         string
}

func ResolvePath(path string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return path, nil
	}
	shortcutsFile := filepath.Join(homeDir, ".anbu", "box-shortcuts.json")
	data, err := os.ReadFile(shortcutsFile)
	if err != nil {
		return path, nil
	}
	var shortcuts map[string]string
	if err := json.Unmarshal(data, &shortcuts); err != nil {
		return path, nil
	}
	result := path
	result = regexp.MustCompile(`%%`).ReplaceAllString(result, "%")
	pattern := regexp.MustCompile(`%([^%]+)%`)
	result = pattern.ReplaceAllStringFunc(result, func(match string) string {
		key := match[1 : len(match)-1]
		if val, ok := shortcuts[key]; ok {
			return val
		}
		return match
	})
	return result, nil
}
```

### Root Directory Trick

Box folder ID `"0"` is a special constant that always refers to the authenticated user's root folder. This is the key pattern used throughout:

```go
func resolvePathToID(client *http.Client, path string, expectedType string) (string, string, error) {
	if path == "" || path == "/" || path == "root" {
		return "0", "folder", nil  // <-- THE TRICK
	}
	// ... walk path segments from root
}
```

---

## 6. Reference Code from ANBU — File & Folder Operations

Source: `anbu/internal/interactions/box/functions.go`

### Path Resolution (resolvePathToID)

```go
func resolvePathToID(client *http.Client, path string, expectedType string) (string, string, error) {
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
		isFirstSegment := (i == 0)
		isLastSegment := (i == len(segments)-1)
		isNumeric := strings.TrimSpace(segment) != "" && isAllDigits(segment)
		if isNumeric {
			if isFirstSegment {
				folderID, folderType, err := tryResolveFolderByID(client, segment)
				if err == nil {
					currentID = folderID
					currentType = folderType
					if isLastSegment {
						if expectedType != "" && currentType != expectedType {
							return "", "", fmt.Errorf("path error: ID '%s' is a %s, but expected a %s", segment, currentType, expectedType)
						}
						return currentID, currentType, nil
					}
					continue
				}
			}
			if isLastSegment {
				fileID, fileType, err := tryResolveFileByID(client, segment)
				if err == nil {
					if expectedType != "" && fileType != expectedType {
						return "", "", fmt.Errorf("path error: ID '%s' is a %s, but expected a %s", segment, fileType, expectedType)
					}
					return fileID, fileType, nil
				}
				folderID, folderType, err := tryResolveFolderByID(client, segment)
				if err == nil {
					if expectedType != "" && folderType != expectedType {
						return "", "", fmt.Errorf("path error: ID '%s' is a %s, but expected a %s", segment, folderType, expectedType)
					}
					return folderID, folderType, nil
				}
			}
		}
		found := false
		offset := 0
		limit := 1000
		for !found {
			req, err := http.NewRequest("GET", fmt.Sprintf(folderItemsURL, currentID), nil)
			if err != nil {
				return "", "", fmt.Errorf("http error creating request: %w", err)
			}
			q := req.URL.Query()
			q.Add("fields", "type,name")
			q.Add("limit", fmt.Sprintf("%d", limit))
			q.Add("offset", fmt.Sprintf("%d", offset))
			req.URL.RawQuery = q.Encode()
			resp, err := client.Do(req)
			if err != nil {
				return "", "", fmt.Errorf("http error listing folder %s: %w", currentID, err)
			}
			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()
				return "", "", fmt.Errorf("api error listing folder %s (status %d): %s", currentID, resp.StatusCode, string(body))
			}
			var items BoxFolderItems
			err = json.NewDecoder(resp.Body).Decode(&items)
			resp.Body.Close()
			if err != nil {
				return "", "", fmt.Errorf("json parse error: %w", err)
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
				return "", "", fmt.Errorf("path error: '%s' is a %s, but expected a %s", segment, currentType, expectedType)
			}
			return currentID, currentType, nil
		}
		if currentType != "folder" {
			return "", "", fmt.Errorf("path error: '%s' in '%s' is a file, not a folder", segment, path)
		}
	}
	return "0", "folder", nil
}
```

### List Folder Contents

```go
func listBoxFolderContents(client *http.Client, folderID string) ([]BoxItemDisplay, []BoxItemDisplay, error) {
	var allFolders []BoxItemDisplay
	var allFiles []BoxItemDisplay
	offset := 0
	limit := 1000
	for {
		req, err := http.NewRequest("GET", fmt.Sprintf(folderItemsURL, folderID), nil)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create request: %v", err)
		}
		q := req.URL.Query()
		q.Add("fields", "type,name,id,size,modified_at")
		q.Add("limit", fmt.Sprintf("%d", limit))
		q.Add("offset", fmt.Sprintf("%d", offset))
		req.URL.RawQuery = q.Encode()
		resp, err := client.Do(req)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to list items: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, nil, handleBoxAPIError("list items", resp)
		}
		var items BoxFolderItems
		err = json.NewDecoder(resp.Body).Decode(&items)
		resp.Body.Close()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse list response: %v", err)
		}
		for _, item := range items.Entries {
			var modTime string
			if item.ModifiedAt != nil {
				t, err := time.Parse(time.RFC3339, *item.ModifiedAt)
				if err == nil {
					modTime = t.Format("2006-01-02 15:04")
				}
			}
			display := BoxItemDisplay{
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
```

### Upload File (Simple — Multipart)

```go
func uploadBoxFile(client *http.Client, localPath string, boxFolderPath string) error {
	parentFolderID := "0"
	if boxFolderPath != "" {
		var err error
		parentFolderID, _, err = resolvePathToID(client, boxFolderPath, "folder")
		if err != nil {
			return fmt.Errorf("failed to find parent folder '%s': %v", boxFolderPath, err)
		}
	}
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file '%s': %v", localPath, err)
	}
	defer file.Close()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fileName := filepath.Base(localPath)
	attributesJSON := fmt.Sprintf(`{"name":"%s", "parent":{"id":"%s"}}`, fileName, parentFolderID)
	if err := writer.WriteField("attributes", attributesJSON); err != nil {
		return fmt.Errorf("failed to write attributes field: %v", err)
	}
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return fmt.Errorf("failed to create form file: %v", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("failed to copy file to form: %v", err)
	}
	writer.Close()
	req, err := http.NewRequest("POST", uploadFileURL, body)
	if err != nil {
		return fmt.Errorf("failed to create upload request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload file: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return handleBoxAPIError("upload file", resp)
	}
	return nil
}
```

### Download File

```go
func downloadBoxFile(client *http.Client, fileID string, boxFilePath string, localPath string) (string, error) {
	if localPath == "" {
		// resolve filename from path or API
		isNumericID := isAllDigits(strings.TrimSpace(boxFilePath))
		if isNumericID {
			req, _ := http.NewRequest("GET", fmt.Sprintf("%s/files/%s", apiBaseURL, fileID), nil)
			q := req.URL.Query()
			q.Add("fields", "name")
			req.URL.RawQuery = q.Encode()
			resp, err := client.Do(req)
			if err == nil {
				defer resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					var file BoxItem
					if json.NewDecoder(resp.Body).Decode(&file) == nil && file.Name != "" {
						localPath = file.Name
					}
				}
			}
		}
		if localPath == "" {
			parts := strings.Split(strings.Trim(boxFilePath, "/"), "/")
			if len(parts) > 0 {
				localPath = parts[len(parts)-1]
			} else {
				localPath = "downloaded_file"
			}
		}
	}
	out, err := os.Create(localPath)
	if err != nil {
		return "", fmt.Errorf("failed to create local file '%s': %v", localPath, err)
	}
	defer out.Close()
	req, _ := http.NewRequest("GET", fmt.Sprintf(fileContentURL, fileID), nil)
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", handleBoxAPIError("download file", resp)
	}
	if _, err := io.Copy(out, resp.Body); err != nil {
		return "", fmt.Errorf("failed to write to local file: %v", err)
	}
	return localPath, nil
}
```

### Upload Folder (Recursive)

```go
func UploadBoxFolder(client *http.Client, localPath string, boxFolderPath string) error {
	parentFolderID := "0"
	if boxFolderPath != "" {
		var err error
		parentFolderID, _, err = resolvePathToID(client, boxFolderPath, "folder")
		if err != nil {
			return fmt.Errorf("failed to find parent folder '%s': %v", boxFolderPath, err)
		}
	}
	rootFolderName := filepath.Base(localPath)
	driveRootFolderID, err := findOrCreateBoxFolder(client, rootFolderName, parentFolderID)
	if err != nil {
		return fmt.Errorf("failed to create root box folder '%s': %v", rootFolderName, err)
	}
	folderIdMap := make(map[string]string)
	folderIdMap[localPath] = driveRootFolderID
	return filepath.WalkDir(localPath, func(currentLocalPath string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if currentLocalPath == localPath {
			return nil
		}
		parentLocalDir := filepath.Dir(currentLocalPath)
		parentBoxId, ok := folderIdMap[parentLocalDir]
		if !ok {
			return fmt.Errorf("could not find parent Box ID for local path: %s", parentLocalDir)
		}
		if d.IsDir() {
			boxFolderId, err := findOrCreateBoxFolder(client, d.Name(), parentBoxId)
			if err != nil {
				return nil // skip on error
			}
			folderIdMap[currentLocalPath] = boxFolderId
		} else {
			// Upload file using multipart POST to uploadFileURL
			// (same pattern as uploadBoxFile but with parentBoxId)
		}
		return nil
	})
}

func findOrCreateBoxFolder(client *http.Client, folderName string, parentId string) (string, error) {
	// 1. List items in parent folder looking for matching name
	// 2. If found, return existing folder ID
	// 3. If not found, POST to /folders to create it
	offset := 0
	limit := 1000
	for {
		req, _ := http.NewRequest("GET", fmt.Sprintf(folderItemsURL, parentId), nil)
		q := req.URL.Query()
		q.Add("fields", "type,name")
		q.Add("limit", fmt.Sprintf("%d", limit))
		q.Add("offset", fmt.Sprintf("%d", offset))
		req.URL.RawQuery = q.Encode()
		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			break
		}
		var items BoxFolderItems
		json.NewDecoder(resp.Body).Decode(&items)
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
	// Create folder
	folderJSON := fmt.Sprintf(`{"name":"%s", "parent":{"id":"%s"}}`, folderName, parentId)
	req, _ := http.NewRequest("POST", uploadFolderURL, bytes.NewBufferString(folderJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return "", handleBoxAPIError("create folder", resp)
	}
	var folder BoxItem
	json.NewDecoder(resp.Body).Decode(&folder)
	return folder.ID, nil
}
```

### Error Handling

```go
func handleBoxAPIError(action string, resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("api request to '%s' failed with status %s (could not read error body)", action, resp.Status)
	}
	var boxErr BoxError
	if json.Unmarshal(body, &boxErr) == nil {
		return fmt.Errorf("api request to '%s' failed: %s - %s", action, boxErr.Code, boxErr.Message)
	}
	return fmt.Errorf("api request to '%s' failed with status %s: %s", action, resp.Status, string(body))
}
```

---

## 7. Reference Code from ANBU — Sync with SHA1 Comparison

Source: `anbu/internal/interactions/box/sync.go`

### Data Structures

```go
type FileTree struct {
	Files map[string]FileInfo
	Dirs  map[string]*FileTree
}

type FileInfo struct {
	Path string
	Hash string
	ID   string
}
```

### Local Hash Computation

```go
func computeLocalHash(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
```

### Remote File Hash (API)

```go
func getRemoteFileHash(client *http.Client, fileID string) (string, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/files/%s", apiBaseURL, fileID), nil)
	q := req.URL.Query()
	q.Add("fields", "sha1")
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var file struct {
		SHA1 string `json:"sha1"`
	}
	json.NewDecoder(resp.Body).Decode(&file)
	return file.SHA1, nil
}
```

### Build Local Tree

```go
func buildLocalTree(rootDir string, ignoreSet map[string]struct{}) (*FileTree, error) {
	tree := &FileTree{
		Files: make(map[string]FileInfo),
		Dirs:  make(map[string]*FileTree),
	}
	err := filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if _, skip := ignoreSet[d.Name()]; skip {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		relPath, _ := filepath.Rel(rootDir, path)
		if relPath == "." {
			return nil
		}
		if d.IsDir() {
			parts := strings.Split(relPath, string(filepath.Separator))
			current := tree
			for _, part := range parts {
				if current.Dirs == nil {
					current.Dirs = make(map[string]*FileTree)
				}
				if _, exists := current.Dirs[part]; !exists {
					current.Dirs[part] = &FileTree{
						Files: make(map[string]FileInfo),
						Dirs:  make(map[string]*FileTree),
					}
				}
				current = current.Dirs[part]
			}
		} else {
			hash, err := computeLocalHash(path)
			if err != nil {
				return nil
			}
			parts := strings.Split(relPath, string(filepath.Separator))
			fileName := parts[len(parts)-1]
			dirPath := strings.Join(parts[:len(parts)-1], string(filepath.Separator))
			current := tree
			if dirPath != "" {
				for _, part := range strings.Split(dirPath, string(filepath.Separator)) {
					if current.Dirs == nil {
						current.Dirs = make(map[string]*FileTree)
					}
					if _, exists := current.Dirs[part]; !exists {
						current.Dirs[part] = &FileTree{
							Files: make(map[string]FileInfo),
							Dirs:  make(map[string]*FileTree),
						}
					}
					current = current.Dirs[part]
				}
			}
			current.Files[fileName] = FileInfo{Path: relPath, Hash: hash}
		}
		return nil
	})
	return tree, err
}
```

### Build Remote Tree

```go
func buildRemoteTree(client *http.Client, folderID string, basePath string, ignoreSet map[string]struct{}) (*FileTree, error) {
	tree := &FileTree{
		Files: make(map[string]FileInfo),
		Dirs:  make(map[string]*FileTree),
	}
	req, _ := http.NewRequest("GET", fmt.Sprintf(folderItemsURL, folderID), nil)
	q := req.URL.Query()
	q.Add("fields", "type,name,id")
	req.URL.RawQuery = q.Encode()
	resp, _ := client.Do(req)
	defer resp.Body.Close()
	var items BoxFolderItems
	json.NewDecoder(resp.Body).Decode(&items)
	for _, item := range items.Entries {
		if _, skip := ignoreSet[item.Name]; skip {
			continue
		}
		itemPath := filepath.Join(basePath, item.Name)
		if item.Type == "folder" {
			subTree, _ := buildRemoteTree(client, item.ID, itemPath, ignoreSet)
			tree.Dirs[item.Name] = subTree
		} else {
			hash, _ := getRemoteFileHash(client, item.ID)
			tree.Files[item.Name] = FileInfo{Path: itemPath, Hash: hash, ID: item.ID}
		}
	}
	return tree, nil
}
```

### Sync Entry Point

```go
func SyncBoxDirectory(client *http.Client, localDir string, remotePath string, concurrency int, ignore []string) error {
	ignoreSet := make(map[string]struct{})
	for _, v := range ignore {
		name := strings.TrimSpace(v)
		if name != "" {
			ignoreSet[name] = struct{}{}
		}
	}
	localTree, err := buildLocalTree(localDir, ignoreSet)
	if err != nil {
		return fmt.Errorf("failed to build local tree: %v", err)
	}
	folderID := "0"
	if remotePath != "" && remotePath != "/" && remotePath != "root" {
		folderID, _, err = resolvePathToID(client, remotePath, "folder")
		if err != nil {
			return fmt.Errorf("failed to resolve remote path: %v", err)
		}
	}
	remoteTree, err := buildRemoteTree(client, folderID, "", ignoreSet)
	if err != nil {
		return fmt.Errorf("failed to build remote tree: %v", err)
	}
	if concurrency < 1 {
		concurrency = 1
	}
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	err = syncTree(client, localTree, remoteTree, localDir, folderID, sem, &wg, 0)
	wg.Wait()
	return err
}
```

### Sync Tree (Core Algorithm)

The `syncTree` function handles:
- **New local files** → upload to remote
- **Modified files** (hash mismatch) → delete remote, upload new version
- **Deleted local files** → delete from remote
- **New remote-only files** → delete from remote (one-way sync)
- **Missing remote folders** → create them
- **Deleted local folders** → delete from remote recursively

Each operation runs in a goroutine with semaphore-based concurrency control.

See full implementation in `anbu/internal/interactions/box/sync.go` lines 336-491.

---

## 8. Reference Code from ANBU — Local Indexing & Search

Source: `anbu/internal/interactions/box/indexing.go`

```go
type IndexItem struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Path         string `json:"path"`
	Type         string `json:"type"`
	Size         int64  `json:"size"`
	ModifiedTime string `json:"modified_time"`
}

type IndexStore struct {
	Provider  string      `json:"provider"`
	RootPath  string      `json:"root_path"`
	Timestamp time.Time   `json:"timestamp"`
	Items     []IndexItem `json:"items"`
}

func GenerateIndex(client *http.Client, rootPath string) error {
	folderID, _, err := resolvePathToID(client, rootPath, "folder")
	if err != nil {
		return err
	}
	var items []IndexItem
	var crawl func(fid, currentPath string) error
	crawl = func(fid, currentPath string) error {
		folders, files, err := listBoxFolderContents(client, fid)
		if err != nil {
			return err
		}
		for _, f := range files {
			fullP := filepath.Join(currentPath, f.Name)
			items = append(items, IndexItem{
				ID: f.ID, Name: f.Name, Path: fullP,
				Type: "file", Size: f.Size, ModifiedTime: f.ModifiedTime,
			})
		}
		for _, f := range folders {
			fullP := filepath.Join(currentPath, f.Name)
			items = append(items, IndexItem{
				ID: f.ID, Name: f.Name, Path: fullP,
				Type: "folder", ModifiedTime: f.ModifiedTime,
			})
			if f.ID != "" {
				if err := crawl(f.ID, fullP); err != nil {
					return err
				}
			}
		}
		return nil
	}
	startPath := rootPath
	if startPath == "" {
		startPath = "/"
	}
	if err := crawl(folderID, startPath); err != nil {
		return err
	}
	return saveIndex(rootPath, items)
}

func SearchIndex(pattern string, searchPath string, excludeDirs, excludeFiles bool) ([]IndexItem, error) {
	idx, err := loadIndex()
	if err != nil {
		return nil, fmt.Errorf("failed to load index: %w", err)
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex: %w", err)
	}
	var results []IndexItem
	for _, item := range idx.Items {
		if excludeDirs && item.Type == "folder" {
			continue
		}
		if excludeFiles && item.Type != "folder" {
			continue
		}
		if re.MatchString(item.Name) {
			results = append(results, item)
		}
	}
	return results, nil
}
```

Index stored at `~/.boxctl/index.json` (or `~/.anbu/box-index.json`).

---

## 9. Reference Code from ANBU — CLI Command Definitions

Source: `anbu/cmd/interactions-cmd/box.go`

```go
var BoxCmd = &cobra.Command{
	Use:   "box",
	Short: "Interact with Box.com to list, upload, download, sync, and index & search files and folders",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if boxFlags.credentialsFile == "" {
			homeDir, _ := os.UserHomeDir()
			boxFlags.credentialsFile = filepath.Join(homeDir, ".anbu", "box-credentials.json")
		}
		if _, err := os.Stat(boxFlags.credentialsFile); os.IsNotExist(err) {
			return fmt.Errorf("credentials file not found at %s", boxFlags.credentialsFile)
		}
		return nil
	},
}

// Subcommands registered:
//   box list [path]              -- list files/folders
//   box upload <local> [remote]  -- upload file/folder
//   box download <remote> [local] -- download file/folder
//   box sync <local> <remote>    -- sync directory (--concurrency, --ignore)
//   box index [path]             -- index for search
//   box search <regex> [path]    -- search index (--exclude-dirs, --exclude-files)
```

---

## 10. New Operations — API Specs

All endpoints use base URL `https://api.box.com/2.0` unless noted.
Upload endpoints use `https://upload.box.com/api/2.0`.

---

### Collaborations

#### POST /collaborations — Create

```
Query: fields (optional), notify (boolean, optional)
```

**Request Body:**
```json
{
  "item": { "type": "file|folder", "id": "STRING" },       // REQUIRED
  "accessible_by": { "type": "user|group", "id": "STRING" }, // REQUIRED (or "login" instead of "id")
  "role": "editor|viewer|previewer|uploader|previewer uploader|viewer uploader|co-owner", // REQUIRED
  "is_access_only": false,
  "can_view_path": false,
  "expires_at": "2025-12-31T23:59:00-07:00"
}
```

**Response 201:**
```json
{
  "id": "STRING",
  "type": "collaboration",
  "item": { "id": "...", "type": "file|folder", "name": "..." },
  "accessible_by": { "id": "...", "type": "user|group", "name": "...", "login": "..." },
  "role": "editor",
  "status": "accepted|pending|rejected",
  "expires_at": null,
  "is_access_only": false,
  "created_by": { "id": "...", "type": "user", "name": "...", "login": "..." },
  "created_at": "2025-01-01T00:00:00-07:00",
  "modified_at": "2025-01-01T00:00:00-07:00",
  "acknowledged_at": null,
  "invite_email": null,
  "acceptance_requirements_status": { ... }
}
```

#### GET /collaborations/{id} — Get

```
Path: collaboration_id (required)
Query: fields (optional)
Response 200: Collaboration object (same as above)
```

#### PUT /collaborations/{id} — Update

```
Path: collaboration_id (required)
Body: { "role": "...", "status": "pending|accepted|rejected", "expires_at": "...", "can_view_path": true }
Response 200: Collaboration object
Response 204: Empty (when role changed to "owner")
```

#### DELETE /collaborations/{id} — Delete

```
Path: collaboration_id (required)
Response 204: Empty
```

#### GET /collaborations — List Pending

```
Query: status="pending" (REQUIRED), fields, offset, limit
Response 200: { "total_count": N, "limit": N, "offset": N, "entries": [Collaboration...] }
```

---

### Comments

#### POST /comments — Create

```json
{
  "message": "STRING",           // REQUIRED
  "tagged_message": "STRING",   // optional, with @[user_id:name] mentions
  "item": { "id": "STRING", "type": "file|comment" }  // REQUIRED
}
```

**Response 201:**
```json
{
  "id": "STRING", "type": "comment",
  "is_reply_comment": false,
  "message": "STRING",
  "tagged_message": "STRING",
  "created_by": { "id": "...", "type": "user", "name": "...", "login": "..." },
  "created_at": "...", "modified_at": "...",
  "item": { "id": "...", "type": "file", "name": "..." }
}
```

#### GET /comments/{id} — Get

```
Path: comment_id (required), Query: fields
Response 200: Comment object
```

#### PUT /comments/{id} — Update

```
Body: { "message": "STRING" }
Response 200: Comment object
```

#### DELETE /comments/{id} — Delete

```
Response 204: Empty
```

---

### Shared Links

#### Add shared link to file — PUT /files/{file_id}

```
Query: fields="shared_link" (REQUIRED)
```

**Request Body:**
```json
{
  "shared_link": {
    "access": "open|company|collaborators",
    "password": "STRING",
    "vanity_name": "STRING",
    "unshared_at": "2025-12-31T23:59:00Z",
    "permissions": {
      "can_download": true,
      "can_preview": true,
      "can_edit": true
    }
  }
}
```

**Response 200:** File object with `shared_link` field populated.

#### Remove shared link — PUT /files/{file_id}

```json
{ "shared_link": null }
```

#### Add shared link to folder — PUT /folders/{folder_id}

Same structure as file shared links.

#### Resolve shared link — GET /shared_items

```
Headers: boxapi: shared_link=URL&shared_link_password=PASSWORD (REQUIRED)
Query: fields
Response 200: File or Folder object
```

---

### Trash Operations

#### GET /folders/trash/items — List trash

```
Query: fields, limit, offset, usemarker, marker, direction (ASC|DESC), sort (name|date|size)
Response 200: { "total_count": N, "limit": N, "offset": N, "entries": [File|Folder|WebLink...] }
```

#### GET /files/{file_id}/trash — Get trashed file

```
Response 200: TrashFile object (includes trashed_at, purged_at, item_status: "trashed")
```

#### GET /folders/{folder_id}/trash — Get trashed folder

```
Response 200: TrashFolder object
```

#### POST /files/{file_id} — Restore file from trash

```
Body (optional): { "name": "new_name", "parent": { "id": "new_parent_id" } }
Response 201: Restored file object
```

#### POST /folders/{folder_id} — Restore folder from trash

```
Body (optional): { "name": "new_name", "parent": { "id": "new_parent_id" } }
Response 201: Restored folder object
```

#### DELETE /files/{file_id}/trash — Permanently delete file

```
Response 204: Empty
```

#### DELETE /folders/{folder_id}/trash — Permanently delete folder

```
Response 204: Empty
```

---

### Search (Server-Side)

#### GET /search

```
Query Parameters:
  query           string    Search term (matches names, descriptions, content, tags)
  scope           string    "user_content" | "enterprise_content"
  file_extensions []string  Filter by extension
  type            string    "file" | "folder" | "web_link"
  ancestor_folder_ids []string  Limit to folders
  content_types   []string  "name" | "description" | "file_content" | "comments" | "tags"
  created_at_range []string From,To RFC3339
  updated_at_range []string From,To RFC3339
  size_range      []int     Min,Max bytes
  owner_user_ids  []string  Filter by owner
  trash_content   string    "non_trashed_only" | "trashed_only" | "all_items"
  sort            string    "modified_at" | "relevance"
  direction       string    "DESC" | "ASC"
  limit           int       Max results
  offset          int       Pagination offset
  fields          []string

Response 200:
{
  "total_count": N,
  "limit": N,
  "offset": N,
  "entries": [ File | Folder | WebLink ... ]
}
```

---

### Users Management

#### GET /users — List users

```
Query: filter_term, user_type (all|managed|external), fields, offset, limit, usemarker, marker
Response 200: { "total_count": N, "entries": [UserFull...] }
```

#### GET /users/{user_id} — Get user

```
Response 200: UserFull object
{
  "id": "STRING", "type": "user",
  "name": "STRING", "login": "email@example.com",
  "created_at": "...", "modified_at": "...",
  "language": "en", "timezone": "America/Los_Angeles",
  "space_amount": 10737418240, "space_used": 629644,
  "max_upload_size": 2147483648,
  "status": "active|inactive|cannot_delete_edit|cannot_delete_edit_upload",
  "job_title": "...", "phone": "...", "address": "...",
  "avatar_url": "...",
  "role": "admin|coadmin|user",
  "is_sync_enabled": true,
  "is_external_collab_restricted": false,
  "enterprise": { "id": "...", "type": "enterprise", "name": "..." },
  "notification_email": { "email": "...", "is_confirmed": true }
}
```

#### POST /users — Create user

```
Body: { "name": "STRING" (REQUIRED), "login": "email" (required unless app user), "role": "coadmin|user", ... }
Response 201: UserFull
```

#### PUT /users/{user_id} — Update user

```
Body: { "name": "...", "login": "...", "role": "...", "status": "...", ... } (all optional)
Response 200: UserFull
```

#### DELETE /users/{user_id} — Delete user

```
Query: notify (bool), force (bool — delete even if user has files)
Response 204: Empty
```

---

### Groups Management

#### GET /groups — List groups

```
Query: filter_term, fields, offset, limit
Response 200: { "total_count": N, "entries": [GroupFull...] }
```

#### GET /groups/{group_id} — Get group

```
Response 200: GroupFull object
{
  "id": "STRING", "type": "group",
  "name": "STRING",
  "group_type": "managed_group|all_users_group",
  "created_at": "...", "modified_at": "...",
  "provenance": "Active Directory",
  "external_sync_identifier": "AD:123456",
  "description": "...",
  "invitability_level": "admins_only|admins_and_members|all_managed_users",
  "member_viewability_level": "admins_only|admins_and_members|all_managed_users",
  "permissions": { "can_invite_as_collaborator": true }
}
```

#### POST /groups — Create group

```
Body: { "name": "STRING" (REQUIRED), "description": "...", "invitability_level": "...", "member_viewability_level": "..." }
Response 201: GroupFull
```

#### PUT /groups/{group_id} — Update group

```
Body: { "name": "...", "description": "...", ... } (all optional)
Response 200: GroupFull
```

#### DELETE /groups/{group_id} — Delete group

```
Response 204: Empty
```

#### GET /groups/{group_id}/memberships — List members

```
Query: offset, limit
Response 200: { "total_count": N, "entries": [GroupMembership...] }

GroupMembership:
{
  "id": "STRING", "type": "group_membership",
  "user": { "id": "...", "type": "user", "name": "...", "login": "..." },
  "group": { "id": "...", "type": "group", "name": "..." },
  "role": "member|admin",
  "created_at": "...", "modified_at": "..."
}
```

#### POST /group_memberships — Add member

```
Body: { "user": { "id": "STRING" }, "group": { "id": "STRING" }, "role": "member|admin" }
Response 201: GroupMembership
```

---

### Chunked Uploads

**Base URL:** `https://upload.box.com/api/2.0`

#### POST /files/upload_sessions — Create session

```
Body: { "folder_id": "STRING", "file_size": INT64, "file_name": "STRING" } (all REQUIRED)
Response 201:
{
  "id": "F971964745A5CD0C001BBE4E58196BFD",
  "type": "upload_session",
  "session_expires_at": "2025-01-01T00:00:00Z",
  "part_size": 8388608,     // MUST use this exact size for each part (except last)
  "total_parts": 13,
  "num_parts_processed": 0,
  "session_endpoints": {
    "upload_part": "https://upload.box.com/api/2.0/files/upload_sessions/.../parts",
    "commit": "https://upload.box.com/api/2.0/files/upload_sessions/.../commit",
    "abort": "https://upload.box.com/api/2.0/files/upload_sessions/...",
    "list_parts": "https://upload.box.com/api/2.0/files/upload_sessions/.../parts",
    "status": "https://upload.box.com/api/2.0/files/upload_sessions/..."
  }
}
```

Also: `POST /files/{file_id}/upload_sessions` for new version of existing file.

#### PUT /files/upload_sessions/{id} — Upload part

```
Headers:
  digest: sha=BASE64_SHA1_OF_CHUNK (REQUIRED)
  content-range: bytes START-END/TOTAL (REQUIRED)
Body: application/octet-stream (raw binary)

Response 200:
{
  "part": {
    "part_id": "STRING",
    "offset": 0,
    "size": 8388608,
    "sha1": "STRING"
  }
}
```

#### POST /files/upload_sessions/{id}/commit — Commit

```
Headers:
  digest: sha=BASE64_SHA1_OF_WHOLE_FILE (REQUIRED)
  if-match: ETAG (optional)
Body: { "parts": [{ "part_id": "...", "offset": N, "size": N, "sha1": "..." }, ...] }

Response 201: { "total_count": 1, "entries": [FileFull] }
Response 202: Accepted (check Retry-After header, poll until 201)
```

#### DELETE /files/upload_sessions/{id} — Abort

```
Response 204: Empty
```

#### GET /files/upload_sessions/{id}/parts — List parts

```
Query: offset, limit
Response 200: { "total_count": N, "entries": [UploadPart...] }
```

---

## 11. Common Error Response Schema

All Box API endpoints return this on error:

```json
{
  "type": "error",
  "status": 400,
  "code": "bad_request|unauthorized|forbidden|not_found|method_not_allowed|conflict|precondition_failed|too_many_requests|internal_server_error|unavailable|item_name_invalid|insufficient_scope|item_name_in_use",
  "message": "Human-readable error description",
  "context_info": { },
  "help_url": "https://developer.box.com/...",
  "request_id": "STRING"
}
```

**Key error codes to handle:**
- `item_name_in_use` — file/folder name conflict (fall back to version update)
- `too_many_requests` — rate limited (check `Retry-After` header)
- `unauthorized` — token expired (refresh and retry)

---

## 12. Option C — OpenAPI Generated Go Client

The Box OpenAPI 3.0 spec has been downloaded to `/Users/tanishqrupaal/box-openapi.json` and a full Go client has been generated at `/Users/tanishqrupaal/box-go-generated/`.

### What was generated

Using `openapi-generator` v7.19.0:

```bash
openapi-generator generate \
  -i /Users/tanishqrupaal/box-openapi.json \
  -g go \
  -o /Users/tanishqrupaal/box-go-generated/ \
  --package-name boxapi
```

**Output:**
- **565 model files** (`model_*.go`) — Go structs for every Box API type
- **73 API service files** (`api_*.go`) — HTTP client methods for every endpoint
- **`client.go`** — `APIClient` struct with all service pointers
- **`configuration.go`** — Config with OAuth2 support via `golang.org/x/oauth2`
- **`go.mod`** — Dependencies: `golang.org/x/oauth2`, `gopkg.in/validator.v2`, `github.com/stretchr/testify`

### How to use it

The generated client already supports OAuth2 via context:

```go
import (
    "context"
    boxapi "github.com/your/boxctl/generated"
    "golang.org/x/oauth2"
)

// Set up OAuth token source (reuse ANBU's auth code)
tokenSource := config.TokenSource(ctx, token)

// Create API client
cfg := boxapi.NewConfiguration()
client := boxapi.NewAPIClient(cfg)

// Inject OAuth token into context
ctx = context.WithValue(context.Background(), boxapi.ContextOAuth2, tokenSource)

// Use typed API calls
collab, _, err := client.CollaborationsAPI.PostCollaborations(ctx).
    PostCollaborationsRequest(boxapi.PostCollaborationsRequest{...}).
    Execute()
```

### Key API services available

```
client.FilesAPI              client.FoldersAPI
client.CollaborationsAPI     client.CommentsAPI
client.UsersAPI              client.GroupsAPI
client.SearchAPI             client.TrashedItemsAPI
client.SharedLinksFilesAPI   client.SharedLinksFoldersAPI
client.UploadsChunkedAPI     client.DownloadsAPI
client.EventsAPI             client.WebhooksAPI
// ... 73 total services
```

### Recommendation for the new tool

**Hybrid approach:**
1. Use ANBU's OAuth/token code as the auth layer (proven, simple)
2. Use the generated structs from `model_*.go` as type definitions (saves manual struct writing)
3. Write your own HTTP client methods (simpler than the generated API layer, more control)
4. Or: use the generated API layer directly for operations you don't need to customize

See `referenceplan_c.md` for the full generated code reference.

---

## 13. Build & Cross-Compilation

```makefile
BINARY_NAME=boxctl
VERSION=$(shell git describe --tags --always --dirty)

.PHONY: build build-all clean

build:
	go build -ldflags "-s -w -X main.version=$(VERSION)" -o $(BINARY_NAME) .

build-all:
	GOOS=linux   GOARCH=amd64 go build -ldflags "-s -w" -o dist/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux   GOARCH=arm64 go build -ldflags "-s -w" -o dist/$(BINARY_NAME)-linux-arm64 .
	GOOS=darwin  GOARCH=amd64 go build -ldflags "-s -w" -o dist/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin  GOARCH=arm64 go build -ldflags "-s -w" -o dist/$(BINARY_NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o dist/$(BINARY_NAME)-windows-amd64.exe .

clean:
	rm -rf dist/ $(BINARY_NAME)
```

**CI integration (GitHub Actions):**

```yaml
- name: Build binaries
  run: make build-all

- name: Upload artifacts
  uses: actions/upload-artifact@v4
  with:
    name: boxctl-binaries
    path: dist/
```

No Node.js, no NPM, no dynamic loading. Pure static Go binaries.
