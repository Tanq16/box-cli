package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tanq16/box/internal/api"
	u "github.com/tanq16/box/internal/utils"
)

var downloadFileID string

var downloadCmd = &cobra.Command{
	Use:   "download <remote> [local]",
	Short: "Download a file or folder from Box",
	Args:  cobra.RangeArgs(0, 2),
	Run: func(cmd *cobra.Command, args []string) {
		var itemID, itemType string
		var remotePath string

		if downloadFileID != "" {
			itemID, itemType = resolveItemByID(downloadFileID)
			if len(args) > 0 {
				remotePath = args[0]
			}
		} else {
			if len(args) == 0 {
				u.PrintFatal("Must specify a remote path or --id", nil)
			}
			remotePath = args[0]
			var err error
			itemID, itemType, err = api.ResolvePath(boxClient, remotePath, "")
			if err != nil {
				u.PrintFatal("Failed to resolve path", err)
			}
		}

		// Determine local path
		localPath := ""
		if downloadFileID != "" && len(args) >= 1 {
			localPath = args[0]
		} else if downloadFileID == "" && len(args) >= 2 {
			localPath = args[1]
		}

		if itemType == "folder" {
			if localPath == "" && remotePath != "" {
				parts := strings.Split(strings.Trim(remotePath, "/"), "/")
				if len(parts) > 0 {
					localPath = parts[len(parts)-1]
				} else {
					localPath = "downloaded_folder"
				}
			}
			if localPath == "" {
				localPath = "downloaded_folder"
			}
			u.PrintInfo(fmt.Sprintf("Downloading folder to '%s'...", localPath))
			if err := api.DownloadFolder(boxClient, itemID, localPath); err != nil {
				u.PrintFatal("Folder download failed", err)
			}
			u.PrintSuccess(fmt.Sprintf("Downloaded to: %s", localPath))
			return
		}

		if localPath == "" && remotePath != "" {
			parts := strings.Split(strings.Trim(remotePath, "/"), "/")
			if len(parts) > 0 {
				localPath = parts[len(parts)-1]
			}
		}

		// Ensure parent directory exists
		if localPath != "" {
			if dir := filepath.Dir(localPath); dir != "." {
				os.MkdirAll(dir, 0755)
			}
		}

		u.PrintInfo("Downloading file...")
		savedPath, err := api.DownloadFile(boxClient, itemID, localPath)
		if err != nil {
			u.PrintFatal("Download failed", err)
		}
		u.PrintSuccess(fmt.Sprintf("Downloaded to: %s", savedPath))
	},
}

func init() {
	downloadCmd.Flags().StringVar(&downloadFileID, "id", "", "Download by file/folder ID instead of path")
	rootCmd.AddCommand(downloadCmd)
}
