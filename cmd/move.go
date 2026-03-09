package cmd

import (
	"fmt"
	"path"

	"github.com/spf13/cobra"
	"github.com/tanq16/box/internal/api"
	u "github.com/tanq16/box/utils"
)

var moveCmd = &cobra.Command{
	Use:   "move <source> <dest>",
	Short: "Move or rename a file or folder on Box",
	Long: `Move a file or folder to a new location, or rename it in place.

If dest is an existing folder, the item is moved into it.
If dest is a path with a new name under an existing parent, the item is moved and renamed.
If source and dest share the same parent, the item is renamed.`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		srcPath := args[0]
		destPath := args[1]

		srcID, srcType, err := api.ResolvePath(boxClient, srcPath, "")
		if err != nil {
			u.PrintFatal("cmd","Failed to resolve source path", err)
		}

		destID, destType, destErr := api.ResolvePath(boxClient, destPath, "")
		if destErr == nil && destType == "folder" {
			item, err := api.MoveItem(boxClient, srcType, srcID, destID)
			if err != nil {
				u.PrintFatal("cmd","Failed to move item", err)
			}
			u.PrintSuccess("cmd",fmt.Sprintf("Moved to: %s (ID: %s)", item.Name, item.ID))
			return
		}

		destParent := path.Dir(destPath)
		destName := path.Base(destPath)

		destParentID, _, err := api.ResolvePath(boxClient, destParent, "folder")
		if err != nil {
			u.PrintFatal("cmd","Failed to resolve destination parent", err)
		}

		srcParent := path.Dir(srcPath)
		srcParentID, _, _ := api.ResolvePath(boxClient, srcParent, "folder")

		if srcParentID == destParentID {
			item, err := api.RenameItem(boxClient, srcType, srcID, destName)
			if err != nil {
				u.PrintFatal("cmd","Failed to rename item", err)
			}
			u.PrintSuccess("cmd",fmt.Sprintf("Renamed to: %s (ID: %s)", item.Name, item.ID))
			return
		}

		item, err := api.MoveItem(boxClient, srcType, srcID, destParentID)
		if err != nil {
			u.PrintFatal("cmd","Failed to move item", err)
		}
		if item.Name != destName {
			item, err = api.RenameItem(boxClient, srcType, srcID, destName)
			if err != nil {
				u.PrintFatal("cmd","Failed to rename item after move", err)
			}
		}
		u.PrintSuccess("cmd",fmt.Sprintf("Moved to: %s (ID: %s)", item.Name, item.ID))
	},
}

func init() {
	rootCmd.AddCommand(moveCmd)
}
