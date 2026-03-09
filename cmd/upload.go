package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tanq16/box/internal/api"
	u "github.com/tanq16/box/utils"
)

var uploadFlags struct {
	chunked bool
}

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
			u.PrintFatal("cmd",fmt.Sprintf("Cannot access '%s'", localPath), err)
		}

		parentFolderID, _, err := api.ResolvePath(boxClient, remotePath, "folder")
		if err != nil {
			u.PrintFatal("cmd",fmt.Sprintf("Failed to resolve remote path '%s'", remotePath), err)
		}

		if info.IsDir() {
			u.PrintInfo("cmd",fmt.Sprintf("Uploading folder '%s' to '%s'...", localPath, remotePath))
			if err := api.UploadFolder(boxClient, localPath, parentFolderID); err != nil {
				u.PrintFatal("cmd","Folder upload failed", err)
			}
			u.PrintSuccess("cmd","Folder upload complete")
			return
		}

		const chunkedThreshold = 50 * 1024 * 1024
		if uploadFlags.chunked || info.Size() > chunkedThreshold {
			u.PrintInfo("cmd",fmt.Sprintf("Uploading '%s' via chunked upload (%s)...", localPath, u.FormatSize(info.Size())))
			if err := api.UploadFileChunked(boxClient, localPath, parentFolderID); err != nil {
				u.PrintFatal("cmd","Chunked upload failed", err)
			}
			u.PrintSuccess("cmd","Upload complete")
			return
		}

		u.PrintInfo("cmd",fmt.Sprintf("Uploading '%s' to '%s'...", localPath, remotePath))
		if err := api.UploadFile(boxClient, localPath, parentFolderID); err != nil {
			u.PrintFatal("cmd","Upload failed", err)
		}
		u.PrintSuccess("cmd","Upload complete")
	},
}

func init() {
	uploadCmd.Flags().BoolVar(&uploadFlags.chunked, "chunked", false, "Force chunked upload")
	rootCmd.AddCommand(uploadCmd)
}
