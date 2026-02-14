package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tanq16/box/internal/api"
)

var downloadFileID string

var downloadCmd = &cobra.Command{
	Use:   "download <remote> [local]",
	Short: "Download a file or folder from Box",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		remotePath := args[0]
		localPath := ""
		if len(args) > 1 {
			localPath = args[1]
		}

		var itemID, itemType string
		var err error

		if downloadFileID != "" {
			itemID = downloadFileID
			// Determine type by trying file first
			info, infoErr := api.GetFileInfo(boxClient, itemID)
			if infoErr == nil {
				itemType = info.Type
			} else {
				itemType = "file" // default assumption
			}
		} else {
			itemID, itemType, err = api.ResolvePath(boxClient, remotePath, "")
			if err != nil {
				return err
			}
		}

		if itemType == "folder" {
			if localPath == "" {
				parts := strings.Split(strings.Trim(remotePath, "/"), "/")
				if len(parts) > 0 {
					localPath = parts[len(parts)-1]
				} else {
					localPath = "downloaded_folder"
				}
			}
			fmt.Fprintf(os.Stderr, "Downloading folder to '%s'...\n", localPath)
			return api.DownloadFolder(boxClient, itemID, localPath)
		}

		if localPath == "" {
			parts := strings.Split(strings.Trim(remotePath, "/"), "/")
			if len(parts) > 0 {
				localPath = parts[len(parts)-1]
			}
		}

		// Ensure parent directory exists
		if dir := filepath.Dir(localPath); dir != "." {
			os.MkdirAll(dir, 0755)
		}

		fmt.Fprintf(os.Stderr, "Downloading file...\n")
		savedPath, err := api.DownloadFile(boxClient, itemID, localPath)
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Downloaded to: %s\n", savedPath)
		return nil
	},
}

func init() {
	downloadCmd.Flags().StringVar(&downloadFileID, "id", "", "Download by file/folder ID instead of path")
	rootCmd.AddCommand(downloadCmd)
}
