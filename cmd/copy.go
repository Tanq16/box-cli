package cmd

import (
	"fmt"
	"path"

	"github.com/spf13/cobra"
	"github.com/tanq16/box/internal/api"
	u "github.com/tanq16/box/internal/utils"
)

var copyName string

var copyCmd = &cobra.Command{
	Use:   "copy <source> <dest-folder>",
	Short: "Copy a file or folder on Box",
	Long: `Copy a file or folder to a destination folder.

Use --name to give the copy a different name.
If dest-folder is a path ending with a new name and the parent exists,
the item is copied there with that name.`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		srcPath := args[0]
		destPath := args[1]

		// Resolve source
		srcID, srcType, err := api.ResolvePath(boxClient, srcPath, "")
		if err != nil {
			u.PrintFatal("Failed to resolve source path", err)
		}

		// Determine destination folder and optional name
		name := copyName

		// Try resolving dest as an existing folder
		destID, destType, destErr := api.ResolvePath(boxClient, destPath, "")
		if destErr == nil && destType == "folder" {
			// Dest is an existing folder — copy into it
			if srcType == "folder" {
				item, err := api.CopyFolder(boxClient, srcID, destID, name)
				if err != nil {
					u.PrintFatal("Failed to copy folder", err)
				}
				u.PrintSuccess(fmt.Sprintf("Copied: %s (ID: %s)", item.Name, item.ID))
			} else {
				item, err := api.CopyFile(boxClient, srcID, destID, name)
				if err != nil {
					u.PrintFatal("Failed to copy file", err)
				}
				u.PrintSuccess(fmt.Sprintf("Copied: %s (ID: %s)", item.Name, item.ID))
			}
			return
		}

		// Dest doesn't exist — treat parent as dest folder, basename as name
		destParent := path.Dir(destPath)
		if name == "" {
			name = path.Base(destPath)
		}

		destParentID, _, err := api.ResolvePath(boxClient, destParent, "folder")
		if err != nil {
			u.PrintFatal("Failed to resolve destination path", err)
		}

		if srcType == "folder" {
			item, err := api.CopyFolder(boxClient, srcID, destParentID, name)
			if err != nil {
				u.PrintFatal("Failed to copy folder", err)
			}
			u.PrintSuccess(fmt.Sprintf("Copied: %s (ID: %s)", item.Name, item.ID))
		} else {
			item, err := api.CopyFile(boxClient, srcID, destParentID, name)
			if err != nil {
				u.PrintFatal("Failed to copy file", err)
			}
			u.PrintSuccess(fmt.Sprintf("Copied: %s (ID: %s)", item.Name, item.ID))
		}
	},
}

func init() {
	copyCmd.Flags().StringVar(&copyName, "name", "", "Name for the copy (defaults to original name)")
	rootCmd.AddCommand(copyCmd)
}
