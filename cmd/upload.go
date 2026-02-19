package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tanq16/box/internal/api"
	u "github.com/tanq16/box/internal/utils"
)

var uploadChunked bool

var uploadCmd = &cobra.Command{
	Use:   "upload <local> [remote]",
	Short: "Upload a file or folder to Box",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		localPath := args[0]
		remotePath := "/"
		if len(args) > 1 {
			remotePath = args[1]
		}

		info, err := os.Stat(localPath)
		if err != nil {
			u.PrintFatal(fmt.Sprintf("Cannot access '%s'", localPath), err)
		}

		parentFolderID, _, err := api.ResolvePath(boxClient, remotePath, "folder")
		if err != nil {
			u.PrintFatal(fmt.Sprintf("Failed to resolve remote path '%s'", remotePath), err)
		}

		if info.IsDir() {
			u.PrintInfo(fmt.Sprintf("Uploading folder '%s' to '%s'...", localPath, remotePath))
			if err := api.UploadFolder(boxClient, localPath, parentFolderID); err != nil {
				u.PrintFatal("Folder upload failed", err)
			}
			u.PrintSuccess("Folder upload complete")
			return
		}

		// Check if chunked upload should be used
		const chunkedThreshold = 50 * 1024 * 1024 // 50MB
		if uploadChunked || info.Size() > chunkedThreshold {
			u.PrintInfo(fmt.Sprintf("Uploading '%s' via chunked upload (%s)...", localPath, u.FormatSize(info.Size())))
			if err := api.UploadFileChunked(boxClient, localPath, parentFolderID); err != nil {
				u.PrintFatal("Chunked upload failed", err)
			}
			u.PrintSuccess("Upload complete")
			return
		}

		u.PrintInfo(fmt.Sprintf("Uploading '%s' to '%s'...", localPath, remotePath))
		if err := api.UploadFile(boxClient, localPath, parentFolderID); err != nil {
			u.PrintFatal("Upload failed", err)
		}
		u.PrintSuccess("Upload complete")
	},
}

func init() {
	uploadCmd.Flags().BoolVar(&uploadChunked, "chunked", false, "Force chunked upload")
	rootCmd.AddCommand(uploadCmd)
}
