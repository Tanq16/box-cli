package cmd

import (
	"fmt"
	"path"

	"github.com/spf13/cobra"
	"github.com/tanq16/box/internal/api"
	u "github.com/tanq16/box/utils"
)

var copyFlags struct {
	name string
}

var copyCmd = &cobra.Command{
	Use:     "copy <source> <dest-folder>",
	Aliases: []string{"cp"},
	Short:   "Copy a file or folder on Box",
	Long: `Copy a file or folder to a destination folder.

Use --name to give the copy a different name.
If dest-folder is a path ending with a new name and the parent exists,
the item is copied there with that name.`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		srcPath := args[0]
		destPath := args[1]

		srcID, srcType, err := api.ResolvePath(boxClient, srcPath, "")
		if err != nil {
			u.PrintFatal("Failed to resolve source path", err)
		}

		name := copyFlags.name

		destID, destType, destErr := api.ResolvePath(boxClient, destPath, "")
		if destErr == nil && destType == "folder" {
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
	copyCmd.Flags().StringVarP(&copyFlags.name, "name", "n", "", "Name for the copy (defaults to original name)")
	rootCmd.AddCommand(copyCmd)
}
