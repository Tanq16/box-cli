package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tanq16/box/internal/api"
	"github.com/tanq16/box/internal/utils"
)

var uploadChunked bool

var uploadCmd = &cobra.Command{
	Use:   "upload <local> [remote]",
	Short: "Upload a file or folder to Box",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		localPath := args[0]
		remotePath := "/"
		if len(args) > 1 {
			remotePath = args[1]
		}

		info, err := os.Stat(localPath)
		if err != nil {
			return fmt.Errorf("cannot access '%s': %w", localPath, err)
		}

		parentFolderID, err := api.ResolveRemoteFolderID(boxClient, remotePath, "")
		if err != nil {
			return fmt.Errorf("failed to resolve remote path '%s': %w", remotePath, err)
		}

		if info.IsDir() {
			fmt.Fprintf(os.Stderr, "Uploading folder '%s' to '%s'...\n", localPath, remotePath)
			return api.UploadFolder(boxClient, localPath, parentFolderID)
		}

		// Check if chunked upload should be used
		const chunkedThreshold = 50 * 1024 * 1024 // 50MB
		if uploadChunked || info.Size() > chunkedThreshold {
			fmt.Fprintf(os.Stderr, "Uploading '%s' via chunked upload (%s)...\n", localPath, utils.FormatSize(info.Size()))
			return api.UploadFileChunked(boxClient, localPath, parentFolderID)
		}

		fmt.Fprintf(os.Stderr, "Uploading '%s' to '%s'...\n", localPath, remotePath)
		if err := api.UploadFile(boxClient, localPath, parentFolderID); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Upload complete.\n")
		return nil
	},
}

func init() {
	uploadCmd.Flags().BoolVar(&uploadChunked, "chunked", false, "Force chunked upload")
	rootCmd.AddCommand(uploadCmd)
}
