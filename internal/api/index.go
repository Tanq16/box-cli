package api

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/tanq16/box/internal/auth"
	"github.com/tanq16/box/internal/client"
	"github.com/tanq16/box/internal/types"
)

func GenerateIndex(c *client.BoxClient, rootPath string) error {
	folderID, _, err := ResolvePath(c, rootPath, "folder")
	if err != nil {
		return err
	}

	var items []types.IndexItem
	var crawl func(fid, currentPath string) error
	crawl = func(fid, currentPath string) error {
		folders, files, err := ListFolder(c, fid)
		if err != nil {
			return err
		}
		for _, f := range files {
			fullPath := filepath.Join(currentPath, f.Name)
			items = append(items, types.IndexItem{
				ID: f.ID, Name: f.Name, Path: fullPath,
				Type: "file", Size: f.Size, ModifiedTime: f.ModifiedTime,
			})
		}
		for _, f := range folders {
			fullPath := filepath.Join(currentPath, f.Name)
			items = append(items, types.IndexItem{
				ID: f.ID, Name: f.Name, Path: fullPath,
				Type: "folder", ModifiedTime: f.ModifiedTime,
			})
			if f.ID != "" {
				if err := crawl(f.ID, fullPath); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: error crawling '%s': %v\n", fullPath, err)
				}
			}
		}
		return nil
	}

	startPath := rootPath
	if startPath == "" || startPath == "/" {
		startPath = "/"
	}
	fmt.Fprintf(os.Stderr, "Indexing from '%s'...\n", startPath)
	if err := crawl(folderID, startPath); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Indexed %d items.\n", len(items))
	return saveIndex(rootPath, items)
}

func saveIndex(rootPath string, items []types.IndexItem) error {
	dir := auth.ConfigDir()
	store := types.IndexStore{
		Provider:  "box",
		RootPath:  rootPath,
		Timestamp: time.Now(),
		Items:     items,
	}
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal index: %w", err)
	}
	p := filepath.Join(dir, "index.json")
	if err := os.WriteFile(p, data, 0644); err != nil {
		return fmt.Errorf("failed to write index: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Index saved to %s\n", p)
	return nil
}

func loadIndex() (*types.IndexStore, error) {
	dir := auth.ConfigDir()
	p := filepath.Join(dir, "index.json")
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("no index found — run 'box index' first: %w", err)
	}
	var store types.IndexStore
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, fmt.Errorf("failed to parse index: %w", err)
	}
	return &store, nil
}

func SearchIndex(pattern string, filterType string, pathPrefix string) ([]types.IndexItem, error) {
	idx, err := loadIndex()
	if err != nil {
		return nil, err
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex: %w", err)
	}
	var results []types.IndexItem
	for _, item := range idx.Items {
		if filterType == "file" && item.Type != "file" {
			continue
		}
		if filterType == "folder" && item.Type != "folder" {
			continue
		}
		if pathPrefix != "" && !strings.HasPrefix(item.Path, pathPrefix) {
			continue
		}
		if re.MatchString(item.Name) || re.MatchString(item.Path) {
			results = append(results, item)
		}
	}
	return results, nil
}
