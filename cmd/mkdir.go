package cmd

import (
	"fmt"
	"path"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tanq16/box/internal/api"
	u "github.com/tanq16/box/utils"
)

var mkdirFlags struct {
	parents bool
}

var mkdirCmd = &cobra.Command{
	Use:   "mkdir <path>",
	Short: "Create a folder on Box",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		remotePath := strings.Trim(args[0], "/")
		if remotePath == "" {
			u.PrintFatal("Must specify a folder path", nil)
		}

		if mkdirFlags.parents {
			segments := strings.Split(remotePath, "/")
			parentID := "0"
			for _, segment := range segments {
				id, err := api.FindOrCreateFolder(boxClient, segment, parentID)
				if err != nil {
					u.PrintFatal(fmt.Sprintf("Failed to create folder '%s'", segment), err)
				}
				parentID = id
			}
			u.PrintSuccess(fmt.Sprintf("Created: /%s", remotePath))
			return
		}

		parentPath := path.Dir(remotePath)
		folderName := path.Base(remotePath)

		parentID, _, err := api.ResolvePath(boxClient, parentPath, "folder")
		if err != nil {
			u.PrintFatal("Failed to resolve parent path", err)
		}

		folder, err := api.CreateFolder(boxClient, folderName, parentID)
		if err != nil {
			u.PrintFatal("Failed to create folder", err)
		}
		u.PrintSuccess(fmt.Sprintf("Created: /%s (ID: %s)", remotePath, folder.ID))
	},
}

func init() {
	mkdirCmd.Flags().BoolVarP(&mkdirFlags.parents, "parents", "p", false, "Create intermediate directories as needed")
	rootCmd.AddCommand(mkdirCmd)
}
