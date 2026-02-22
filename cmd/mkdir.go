package cmd

import (
	"fmt"
	"path"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tanq16/box/internal/api"
	u "github.com/tanq16/box/internal/utils"
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
			u.PrintFatal("cmd","Must specify a folder path", nil)
		}

		if mkdirFlags.parents {
			// Create all intermediate directories
			segments := strings.Split(remotePath, "/")
			parentID := "0"
			for _, segment := range segments {
				id, err := api.FindOrCreateFolder(boxClient, segment, parentID)
				if err != nil {
					u.PrintFatal("cmd",fmt.Sprintf("Failed to create folder '%s'", segment), err)
				}
				parentID = id
			}
			u.PrintSuccess("cmd",fmt.Sprintf("Created: /%s", remotePath))
			return
		}

		// Single folder creation
		parentPath := path.Dir(remotePath)
		folderName := path.Base(remotePath)

		parentID, _, err := api.ResolvePath(boxClient, parentPath, "folder")
		if err != nil {
			u.PrintFatal("cmd","Failed to resolve parent path", err)
		}

		folder, err := api.CreateFolder(boxClient, folderName, parentID)
		if err != nil {
			u.PrintFatal("cmd","Failed to create folder", err)
		}
		u.PrintSuccess("cmd",fmt.Sprintf("Created: /%s (ID: %s)", remotePath, folder.ID))
	},
}

func init() {
	mkdirCmd.Flags().BoolVarP(&mkdirFlags.parents, "parents", "p", false, "Create intermediate directories as needed")
	rootCmd.AddCommand(mkdirCmd)
}
