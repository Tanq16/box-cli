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
	"sync"
	"sync/atomic"

	"github.com/rs/zerolog/log"
	"github.com/tanq16/box/internal/client"
	"github.com/tanq16/box/internal/types"
	"golang.org/x/sync/errgroup"
)

type SyncAction int

const (
	SyncAdd SyncAction = iota
	SyncUpdate
	SyncDelete
	SyncCreateFolder
	SyncDeleteFolder
)

type SyncOp struct {
	Action SyncAction
	Path   string
}

type SyncPlan struct {
	Add     int // new files to upload (push) or download (pull)
	Update  int
	Delete  int
	Folders int
	Total   int
	Ops     []SyncOp

	localTree      *types.FileTree
	remoteTree     *types.FileTree
	localDir       string
	remoteFolderID string
}

func (p *SyncPlan) DeleteTotal() int {
	n := 0
	for _, op := range p.Ops {
		if op.Action == SyncDelete || op.Action == SyncDeleteFolder {
			n++
		}
	}
	return n
}

func (p *SyncPlan) HasDeletes() bool {
	return p.DeleteTotal() > 0
}

type SyncFailure struct {
	Item string
	Err  error
}

type SyncProgress struct {
	Completed atomic.Int32
	Errors    atomic.Int32

	mu       sync.Mutex
	Failures []SyncFailure
}

func (p *SyncProgress) fail(item string, err error) {
	p.Errors.Add(1)
	p.mu.Lock()
	p.Failures = append(p.Failures, SyncFailure{Item: item, Err: err})
	p.mu.Unlock()
}

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
				log.Debug().Err(err).Str("file", path).Msg("failed to compute hash, skipping")
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

func BuildRemoteTree(c *client.BoxClient, folderID string, basePath string, ignoreSet map[string]struct{}, localHint *types.FileTree) (*types.FileTree, error) {
	tree := &types.FileTree{
		ID:    folderID,
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
		q.Add("fields", "type,name,id,sha1")
		q.Add("limit", fmt.Sprintf("%d", limit))
		q.Add("offset", fmt.Sprintf("%d", offset))
		req.URL.RawQuery = q.Encode()

		resp, err := c.Do(req)
		if err != nil {
			return nil, err
		}
		var items types.BoxFolderItems
		if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}
		resp.Body.Close()

		for _, item := range items.Entries {
			if _, skip := ignoreSet[item.Name]; skip {
				continue
			}
			itemPath := filepath.Join(basePath, item.Name)
			if item.Type == "folder" {
				var childHint *types.FileTree
				if localHint != nil {
					sub, ok := localHint.Dirs[item.Name]
					if !ok {
						tree.Dirs[item.Name] = &types.FileTree{
							ID:    item.ID,
							Files: make(map[string]types.FileInfo),
							Dirs:  make(map[string]*types.FileTree),
						}
						continue
					}
					childHint = sub
				}
				subTree, err := BuildRemoteTree(c, item.ID, itemPath, ignoreSet, childHint)
				if err != nil {
					log.Debug().Err(err).Str("folder", item.Name).Msg("failed to build remote subtree, skipping")
					continue
				}
				tree.Dirs[item.Name] = subTree
			} else {
				tree.Files[item.Name] = types.FileInfo{Path: itemPath, Hash: item.SHA1, ID: item.ID}
			}
		}
		offset += len(items.Entries)
		if offset >= items.TotalCount || len(items.Entries) == 0 {
			break
		}
	}
	return tree, nil
}

func collectPushOps(local, remote *types.FileTree, prefix string) []SyncOp {
	var ops []SyncOp
	for name, localFile := range local.Files {
		remoteFile, exists := remote.Files[name]
		if !exists {
			ops = append(ops, SyncOp{SyncAdd, filepath.Join(prefix, name)})
		} else if remoteFile.Hash != localFile.Hash {
			ops = append(ops, SyncOp{SyncUpdate, filepath.Join(prefix, name)})
		}
	}
	for name := range remote.Files {
		if _, exists := local.Files[name]; !exists {
			ops = append(ops, SyncOp{SyncDelete, filepath.Join(prefix, name)})
		}
	}
	for name, localSub := range local.Dirs {
		remoteSub, exists := remote.Dirs[name]
		if !exists {
			ops = append(ops, SyncOp{SyncCreateFolder, filepath.Join(prefix, name)})
			remoteSub = &types.FileTree{
				Files: make(map[string]types.FileInfo),
				Dirs:  make(map[string]*types.FileTree),
			}
		}
		ops = append(ops, collectPushOps(localSub, remoteSub, filepath.Join(prefix, name))...)
	}
	for name := range remote.Dirs {
		if _, exists := local.Dirs[name]; !exists {
			ops = append(ops, SyncOp{SyncDeleteFolder, filepath.Join(prefix, name)})
		}
	}
	return ops
}

func collectPullOps(remote, local *types.FileTree, prefix string) []SyncOp {
	var ops []SyncOp
	for name, remoteFile := range remote.Files {
		localFile, exists := local.Files[name]
		if !exists {
			ops = append(ops, SyncOp{SyncAdd, filepath.Join(prefix, name)})
		} else if localFile.Hash != remoteFile.Hash {
			ops = append(ops, SyncOp{SyncUpdate, filepath.Join(prefix, name)})
		}
	}
	for name := range local.Files {
		if _, exists := remote.Files[name]; !exists {
			ops = append(ops, SyncOp{SyncDelete, filepath.Join(prefix, name)})
		}
	}
	for name, remoteSub := range remote.Dirs {
		localSub, exists := local.Dirs[name]
		if !exists {
			localSub = &types.FileTree{
				Files: make(map[string]types.FileInfo),
				Dirs:  make(map[string]*types.FileTree),
			}
		}
		ops = append(ops, collectPullOps(remoteSub, localSub, filepath.Join(prefix, name))...)
	}
	for name := range local.Dirs {
		if _, exists := remote.Dirs[name]; !exists {
			ops = append(ops, SyncOp{SyncDeleteFolder, filepath.Join(prefix, name)})
		}
	}
	return ops
}

func summarize(ops []SyncOp) (add, update, del, folders int) {
	for _, op := range ops {
		switch op.Action {
		case SyncAdd:
			add++
		case SyncUpdate:
			update++
		case SyncDelete:
			del++
		case SyncCreateFolder, SyncDeleteFolder:
			folders++
		}
	}
	return
}

func PlanPush(ctx context.Context, c *client.BoxClient, localDir string, remotePath string, ignore []string) (*SyncPlan, error) {
	ignoreSet := makeIgnoreSet(ignore)

	localTree, err := BuildLocalTree(localDir, ignoreSet)
	if err != nil {
		return nil, fmt.Errorf("failed to build local tree: %w", err)
	}

	folderID, err := ResolveRemoteFolderID(c, remotePath, "")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve remote path: %w", err)
	}

	remoteTree, err := BuildRemoteTree(c, folderID, "", ignoreSet, localTree)
	if err != nil {
		return nil, fmt.Errorf("failed to build remote tree: %w", err)
	}

	ops := collectPushOps(localTree, remoteTree, "")
	add, update, del, folders := summarize(ops)

	return &SyncPlan{
		Add:            add,
		Update:         update,
		Delete:         del,
		Folders:        folders,
		Total:          add + update + del + folders,
		Ops:            ops,
		localTree:      localTree,
		remoteTree:     remoteTree,
		localDir:       localDir,
		remoteFolderID: folderID,
	}, nil
}

func ExecPush(ctx context.Context, c *client.BoxClient, plan *SyncPlan, concurrency int, progress *SyncProgress, deleteEnabled bool) error {
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(concurrency)

	execPushTree(ctx, g, c, plan.localTree, plan.remoteTree, plan.localDir, plan.remoteFolderID, progress, deleteEnabled)
	return g.Wait()
}

func execPushTree(ctx context.Context, g *errgroup.Group, c *client.BoxClient, local *types.FileTree, remote *types.FileTree, localDir string, remoteFolderID string, progress *SyncProgress, deleteEnabled bool) {
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
						progress.fail(name, err)
					}
				} else {
					if err := UploadFile(c, localPath, remoteFolderID, true, nil); err != nil {
						progress.fail(name, err)
					}
				}
				progress.Completed.Add(1)
				return nil
			})
		}
	}

	if deleteEnabled {
		for name, remoteFile := range remote.Files {
			if _, exists := local.Files[name]; !exists {
				g.Go(func() error {
					if ctx.Err() != nil {
						return ctx.Err()
					}
					if err := DeleteFile(c, remoteFile.ID); err != nil {
						progress.fail(name, err)
					}
					progress.Completed.Add(1)
					return nil
				})
			}
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
				progress.fail(name, err)
				progress.Completed.Add(1)
				continue
			}
			subFolderID = id
			remoteSubTree = &types.FileTree{
				Files: make(map[string]types.FileInfo),
				Dirs:  make(map[string]*types.FileTree),
			}
			progress.Completed.Add(1)
		} else {
			subFolderID = remoteSubTree.ID
		}

		execPushTree(ctx, g, c, localSubTree, remoteSubTree, subLocalDir, subFolderID, progress, deleteEnabled)
	}

	if deleteEnabled {
		for name := range remote.Dirs {
			if _, exists := local.Dirs[name]; !exists {
				folderID := remote.Dirs[name].ID
				if folderID != "" {
					if err := DeleteFolder(c, folderID, true); err != nil {
						progress.fail(name, err)
					}
					progress.Completed.Add(1)
				}
			}
		}
	}
}

func PlanPull(ctx context.Context, c *client.BoxClient, remotePath string, localDir string, ignore []string) (*SyncPlan, error) {
	ignoreSet := makeIgnoreSet(ignore)

	folderID, err := ResolveRemoteFolderID(c, remotePath, "")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve remote path: %w", err)
	}

	remoteTree, err := BuildRemoteTree(c, folderID, "", ignoreSet, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build remote tree: %w", err)
	}

	localTree, err := BuildLocalTree(localDir, ignoreSet)
	if err != nil {
		os.MkdirAll(localDir, 0755)
		localTree = &types.FileTree{
			Files: make(map[string]types.FileInfo),
			Dirs:  make(map[string]*types.FileTree),
		}
	}

	ops := collectPullOps(remoteTree, localTree, "")
	add, update, del, folders := summarize(ops)

	return &SyncPlan{
		Add:            add,
		Update:         update,
		Delete:         del,
		Folders:        folders,
		Total:          add + update + del + folders,
		Ops:            ops,
		localTree:      localTree,
		remoteTree:     remoteTree,
		localDir:       localDir,
		remoteFolderID: folderID,
	}, nil
}

func ExecPull(ctx context.Context, c *client.BoxClient, plan *SyncPlan, concurrency int, progress *SyncProgress, deleteEnabled bool) error {
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(concurrency)

	execPullTree(ctx, g, c, plan.remoteTree, plan.localTree, plan.localDir, plan.remoteFolderID, progress, deleteEnabled)
	return g.Wait()
}

func execPullTree(ctx context.Context, g *errgroup.Group, c *client.BoxClient, remote *types.FileTree, local *types.FileTree, localDir string, remoteFolderID string, progress *SyncProgress, deleteEnabled bool) {
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
					progress.fail(name, err)
				}
				progress.Completed.Add(1)
				return nil
			})
		}
	}

	if deleteEnabled {
		for name := range local.Files {
			if _, exists := remote.Files[name]; !exists {
				localPath := filepath.Join(localDir, name)
				if err := os.Remove(localPath); err != nil {
					progress.fail(name, err)
				}
				progress.Completed.Add(1)
			}
		}
	}

	for name, remoteSubTree := range remote.Dirs {
		if ctx.Err() != nil {
			return
		}
		localSubTree, exists := local.Dirs[name]
		subLocalDir := filepath.Join(localDir, name)
		subFolderID := remoteSubTree.ID

		if !exists {
			localSubTree = &types.FileTree{
				Files: make(map[string]types.FileInfo),
				Dirs:  make(map[string]*types.FileTree),
			}
		}

		execPullTree(ctx, g, c, remoteSubTree, localSubTree, subLocalDir, subFolderID, progress, deleteEnabled)
	}

	if deleteEnabled {
		for name := range local.Dirs {
			if _, exists := remote.Dirs[name]; !exists {
				localPath := filepath.Join(localDir, name)
				if err := os.RemoveAll(localPath); err != nil {
					progress.fail(name, err)
				}
				progress.Completed.Add(1)
			}
		}
	}
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
