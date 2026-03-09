package api

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/tanq16/box/internal/client"
	"github.com/tanq16/box/internal/types"
	"golang.org/x/sync/errgroup"
)

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

func getRemoteFileHash(c *client.BoxClient, fileID string) (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/files/%s", client.APIBaseURL, fileID), nil)
	if err != nil {
		return "", err
	}
	q := req.URL.Query()
	q.Add("fields", "sha1")
	req.URL.RawQuery = q.Encode()

	var file struct {
		SHA1 string `json:"sha1"`
	}
	_, err = c.DoJSON(req, &file)
	return file.SHA1, err
}

func BuildLocalTree(rootDir string, ignoreSet map[string]struct{}) (*types.FileTree, error) {
	tree := &types.FileTree{
		Files: make(map[string]types.FileInfo),
		Dirs:  make(map[string]*types.FileTree),
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
					current.Dirs = make(map[string]*types.FileTree)
				}
				if _, exists := current.Dirs[part]; !exists {
					current.Dirs[part] = &types.FileTree{
						Files: make(map[string]types.FileInfo),
						Dirs:  make(map[string]*types.FileTree),
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
			dirParts := parts[:len(parts)-1]
			current := tree
			for _, part := range dirParts {
				if current.Dirs == nil {
					current.Dirs = make(map[string]*types.FileTree)
				}
				if _, exists := current.Dirs[part]; !exists {
					current.Dirs[part] = &types.FileTree{
						Files: make(map[string]types.FileInfo),
						Dirs:  make(map[string]*types.FileTree),
					}
				}
				current = current.Dirs[part]
			}
			current.Files[fileName] = types.FileInfo{Path: relPath, Hash: hash}
		}
		return nil
	})
	return tree, err
}

func BuildRemoteTree(c *client.BoxClient, folderID string, basePath string, ignoreSet map[string]struct{}) (*types.FileTree, error) {
	tree := &types.FileTree{
		Files: make(map[string]types.FileInfo),
		Dirs:  make(map[string]*types.FileTree),
	}

	offset := 0
	limit := 1000
	for {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/folders/%s/items", client.APIBaseURL, folderID), nil)
		if err != nil {
			return nil, err
		}
		q := req.URL.Query()
		q.Add("fields", "type,name,id")
		q.Add("limit", fmt.Sprintf("%d", limit))
		q.Add("offset", fmt.Sprintf("%d", offset))
		req.URL.RawQuery = q.Encode()

		resp, err := c.Do(req)
		if err != nil {
			return nil, err
		}
		var items types.BoxFolderItems
		json.NewDecoder(resp.Body).Decode(&items)
		resp.Body.Close()

		for _, item := range items.Entries {
			if _, skip := ignoreSet[item.Name]; skip {
				continue
			}
			itemPath := filepath.Join(basePath, item.Name)
			if item.Type == "folder" {
				subTree, err := BuildRemoteTree(c, item.ID, itemPath, ignoreSet)
				if err != nil {
					continue
				}
				tree.Dirs[item.Name] = subTree
			} else {
				hash, _ := getRemoteFileHash(c, item.ID)
				tree.Files[item.Name] = types.FileInfo{Path: itemPath, Hash: hash, ID: item.ID}
			}
		}
		offset += len(items.Entries)
		if offset >= items.TotalCount || len(items.Entries) == 0 {
			break
		}
	}
	return tree, nil
}

func SyncPush(ctx context.Context, c *client.BoxClient, localDir string, remotePath string, concurrency int, ignore []string) error {
	ignoreSet := makeIgnoreSet(ignore)

	localTree, err := BuildLocalTree(localDir, ignoreSet)
	if err != nil {
		return fmt.Errorf("failed to build local tree: %w", err)
	}

	folderID, err := ResolveRemoteFolderID(c, remotePath, "")
	if err != nil {
		return fmt.Errorf("failed to resolve remote path: %w", err)
	}

	remoteTree, err := BuildRemoteTree(c, folderID, "", ignoreSet)
	if err != nil {
		return fmt.Errorf("failed to build remote tree: %w", err)
	}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(concurrency)

	syncPushTree(ctx, g, c, localTree, remoteTree, localDir, folderID)
	return g.Wait()
}

func syncPushTree(ctx context.Context, g *errgroup.Group, c *client.BoxClient, local *types.FileTree, remote *types.FileTree, localDir string, remoteFolderID string) {
	for name, localFile := range local.Files {
		remoteFile, exists := remote.Files[name]
		if !exists || remoteFile.Hash != localFile.Hash {
			g.Go(func() error {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				localPath := filepath.Join(localDir, name)
				if exists && remoteFile.ID != "" {
					if err := UploadFileVersion(c, localPath, remoteFile.ID); err != nil {
						fmt.Fprintf(os.Stderr, "Error updating '%s': %v\n", name, err)
					} else {
						fmt.Fprintf(os.Stderr, "Updated: %s\n", name)
					}
				} else {
					if err := UploadFile(c, localPath, remoteFolderID); err != nil {
						fmt.Fprintf(os.Stderr, "Error uploading '%s': %v\n", name, err)
					} else {
						fmt.Fprintf(os.Stderr, "Uploaded: %s\n", name)
					}
				}
				return nil
			})
		}
	}

	for name, remoteFile := range remote.Files {
		if _, exists := local.Files[name]; !exists {
			g.Go(func() error {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				if err := DeleteFile(c, remoteFile.ID); err != nil {
					fmt.Fprintf(os.Stderr, "Error deleting remote file '%s': %v\n", name, err)
				} else {
					fmt.Fprintf(os.Stderr, "Deleted remote: %s\n", name)
				}
				return nil
			})
		}
	}

	for name, localSubTree := range local.Dirs {
		if ctx.Err() != nil {
			return
		}
		remoteSubTree, exists := remote.Dirs[name]
		subLocalDir := filepath.Join(localDir, name)

		var subFolderID string
		if !exists {
			id, err := FindOrCreateFolder(c, name, remoteFolderID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating folder '%s': %v\n", name, err)
				continue
			}
			subFolderID = id
			remoteSubTree = &types.FileTree{
				Files: make(map[string]types.FileInfo),
				Dirs:  make(map[string]*types.FileTree),
			}
			fmt.Fprintf(os.Stderr, "Created folder: %s\n", name)
		} else {
			subFolderID = getFolderIDFromRemote(c, remoteFolderID, name)
		}

		syncPushTree(ctx, g, c, localSubTree, remoteSubTree, subLocalDir, subFolderID)
	}

	for name := range remote.Dirs {
		if _, exists := local.Dirs[name]; !exists {
			folderID := getFolderIDFromRemote(c, remoteFolderID, name)
			if folderID != "" {
				if err := DeleteFolder(c, folderID); err != nil {
					fmt.Fprintf(os.Stderr, "Error deleting remote folder '%s': %v\n", name, err)
				} else {
					fmt.Fprintf(os.Stderr, "Deleted remote folder: %s\n", name)
				}
			}
		}
	}
}

func SyncPull(ctx context.Context, c *client.BoxClient, remotePath string, localDir string, concurrency int, ignore []string) error {
	ignoreSet := makeIgnoreSet(ignore)

	folderID, err := ResolveRemoteFolderID(c, remotePath, "")
	if err != nil {
		return fmt.Errorf("failed to resolve remote path: %w", err)
	}

	remoteTree, err := BuildRemoteTree(c, folderID, "", ignoreSet)
	if err != nil {
		return fmt.Errorf("failed to build remote tree: %w", err)
	}

	localTree, err := BuildLocalTree(localDir, ignoreSet)
	if err != nil {
		os.MkdirAll(localDir, 0755)
		localTree = &types.FileTree{
			Files: make(map[string]types.FileInfo),
			Dirs:  make(map[string]*types.FileTree),
		}
	}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(concurrency)

	syncPullTree(ctx, g, c, remoteTree, localTree, localDir, folderID)
	return g.Wait()
}

func syncPullTree(ctx context.Context, g *errgroup.Group, c *client.BoxClient, remote *types.FileTree, local *types.FileTree, localDir string, remoteFolderID string) {
	os.MkdirAll(localDir, 0755)

	for name, remoteFile := range remote.Files {
		localFile, exists := local.Files[name]
		if !exists || localFile.Hash != remoteFile.Hash {
			g.Go(func() error {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				localPath := filepath.Join(localDir, name)
				if _, err := DownloadFile(c, remoteFile.ID, localPath); err != nil {
					fmt.Fprintf(os.Stderr, "Error downloading '%s': %v\n", name, err)
				} else {
					fmt.Fprintf(os.Stderr, "Downloaded: %s\n", name)
				}
				return nil
			})
		}
	}

	for name := range local.Files {
		if _, exists := remote.Files[name]; !exists {
			localPath := filepath.Join(localDir, name)
			if err := os.Remove(localPath); err != nil {
				fmt.Fprintf(os.Stderr, "Error deleting local file '%s': %v\n", name, err)
			} else {
				fmt.Fprintf(os.Stderr, "Deleted local: %s\n", name)
			}
		}
	}

	for name, remoteSubTree := range remote.Dirs {
		if ctx.Err() != nil {
			return
		}
		localSubTree, exists := local.Dirs[name]
		subLocalDir := filepath.Join(localDir, name)
		subFolderID := getFolderIDFromRemote(c, remoteFolderID, name)

		if !exists {
			localSubTree = &types.FileTree{
				Files: make(map[string]types.FileInfo),
				Dirs:  make(map[string]*types.FileTree),
			}
		}

		syncPullTree(ctx, g, c, remoteSubTree, localSubTree, subLocalDir, subFolderID)
	}

	for name := range local.Dirs {
		if _, exists := remote.Dirs[name]; !exists {
			localPath := filepath.Join(localDir, name)
			if err := os.RemoveAll(localPath); err != nil {
				fmt.Fprintf(os.Stderr, "Error deleting local folder '%s': %v\n", name, err)
			} else {
				fmt.Fprintf(os.Stderr, "Deleted local folder: %s\n", name)
			}
		}
	}
}

func getFolderIDFromRemote(c *client.BoxClient, parentFolderID string, folderName string) string {
	offset := 0
	limit := 1000
	for {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/folders/%s/items", client.APIBaseURL, parentFolderID), nil)
		if err != nil {
			return ""
		}
		q := req.URL.Query()
		q.Add("fields", "type,name,id")
		q.Add("limit", fmt.Sprintf("%d", limit))
		q.Add("offset", fmt.Sprintf("%d", offset))
		req.URL.RawQuery = q.Encode()

		resp, err := c.Do(req)
		if err != nil {
			return ""
		}
		var items types.BoxFolderItems
		json.NewDecoder(resp.Body).Decode(&items)
		resp.Body.Close()

		for _, item := range items.Entries {
			if strings.EqualFold(item.Name, folderName) && item.Type == "folder" {
				return item.ID
			}
		}
		offset += len(items.Entries)
		if offset >= items.TotalCount || len(items.Entries) == 0 {
			break
		}
	}
	return ""
}

func makeIgnoreSet(ignore []string) map[string]struct{} {
	set := make(map[string]struct{})
	for _, v := range ignore {
		name := strings.TrimSpace(v)
		if name != "" {
			set[name] = struct{}{}
		}
	}
	return set
}
