package cmd

import (
	"errors"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/spf13/cobra"
	"github.com/tanq16/box/internal/api"
	u "github.com/tanq16/box/utils"
)

var uploadFlags struct {
	chunked   bool
	overwrite bool
}

var uploadCmd = &cobra.Command{
	Use:   "upload <local> [remote]",
	Short: "Upload a file or folder to Box",
	Long: `Upload a file or folder to Box.

By default, the upload fails if a file with the same name already exists at the
destination. Use --overwrite to upload as a new version of the existing file.`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		localPath := args[0]
		remotePath := "/"
		if len(args) > 1 {
			remotePath = args[1]
		}

		info, err := os.Stat(localPath)
		if err != nil {
			u.PrintFatal("cmd", fmt.Sprintf("Cannot access '%s'", localPath), err)
		}

		parentFolderID, _, err := api.ResolvePath(boxClient, remotePath, "folder")
		if err != nil {
			u.PrintFatal("cmd", fmt.Sprintf("Failed to resolve remote path '%s'", remotePath), err)
		}

		if info.IsDir() {
			runFolderUpload(localPath, remotePath, parentFolderID)
			return
		}

		const chunkedThreshold = 50 * 1024 * 1024
		if uploadFlags.chunked || info.Size() > chunkedThreshold {
			runChunkedUpload(localPath, info.Size(), parentFolderID)
			return
		}

		runSingleUpload(localPath, info.Size(), remotePath, parentFolderID)
	},
}

func runSingleUpload(localPath string, size int64, remotePath string, parentFolderID string) {
	progress := &api.UploadProgress{Total: size}
	label := fmt.Sprintf("Uploading '%s' to '%s'", localPath, remotePath)

	u.PrintRunning("cmd", label)
	stop, printed := startUploadTicker(progress)

	err := api.UploadFile(boxClient, localPath, parentFolderID, uploadFlags.overwrite, progress)

	close(stop)
	if printed.Load() {
		u.ClearPreviousLine()
	}
	u.ClearLines(1)

	handleUploadResult(err, "Upload complete")
}

func runChunkedUpload(localPath string, size int64, parentFolderID string) {
	progress := &api.UploadProgress{Total: size}
	label := fmt.Sprintf("Uploading '%s' via chunked upload (%s)", localPath, u.FormatSize(size))

	u.PrintRunning("cmd", label)
	stop, printed := startUploadTicker(progress)

	err := api.UploadFileChunked(boxClient, localPath, parentFolderID, uploadFlags.overwrite, progress)

	close(stop)
	if printed.Load() {
		u.ClearPreviousLine()
	}
	u.ClearLines(1)

	handleUploadResult(err, "Upload complete")
}

func runFolderUpload(localPath, remotePath, parentFolderID string) {
	total, err := api.SumFolderSize(localPath)
	if err != nil {
		u.PrintFatal("cmd", fmt.Sprintf("Failed to scan '%s'", localPath), err)
	}
	progress := &api.UploadProgress{Total: total}
	label := fmt.Sprintf("Uploading folder '%s' to '%s' (%s)", localPath, remotePath, u.FormatSize(total))

	u.PrintRunning("cmd", label)
	stop, printed := startUploadTicker(progress)

	err = api.UploadFolder(boxClient, localPath, parentFolderID, uploadFlags.overwrite, progress)

	close(stop)
	if printed.Load() {
		u.ClearPreviousLine()
	}
	u.ClearLines(1)

	handleUploadResult(err, "Folder upload complete")
}

func startUploadTicker(progress *api.UploadProgress) (chan struct{}, *atomic.Bool) {
	stop := make(chan struct{})
	printed := &atomic.Bool{}
	if progress.Total <= 0 {
		return stop, printed
	}
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		firstTick := true
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				done := progress.BytesDone.Load()
				if done > progress.Total {
					done = progress.Total
				}
				pct := int(done * 100 / progress.Total)
				if !firstTick {
					u.ClearPreviousLine()
				}
				firstTick = false
				printed.Store(true)
				u.PrintProgress(fmt.Sprintf("Uploading (%s / %s)", u.FormatSize(done), u.FormatSize(progress.Total)), pct)
			}
		}
	}()
	return stop, printed
}

func handleUploadResult(err error, successMsg string) {
	if err != nil {
		var conflict *api.ConflictError
		if errors.As(err, &conflict) {
			u.PrintFatal("cmd", err.Error(), nil)
		}
		u.PrintFatal("cmd", "Upload failed", err)
	}
	u.PrintSuccess("cmd", successMsg)
}

func init() {
	uploadCmd.Flags().BoolVar(&uploadFlags.chunked, "chunked", false, "Force chunked upload")
	uploadCmd.Flags().BoolVar(&uploadFlags.overwrite, "overwrite", false, "Upload as a new version if a file with the same name exists")
	rootCmd.AddCommand(uploadCmd)
}
