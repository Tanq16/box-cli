package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tanq16/box/cmd/cmdutil"
	"github.com/tanq16/box/internal/api"
	u "github.com/tanq16/box/utils"
)

var deleteFlags struct {
	itemID    string
	recursive bool
}

var deleteCmd = &cobra.Command{
	Use:     "delete <path>",
	Aliases: []string{"rm"},
	Short:   "Move a file or folder to the Box trash",
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		itemID, itemType := cmdutil.ResolveItem(args, deleteFlags.itemID)

		if itemType == "folder" {
			if err := api.DeleteFolder(boxClient, itemID, deleteFlags.recursive); err != nil {
				u.PrintFatal("Failed to delete folder", err)
			}
		} else {
			if err := api.DeleteFile(boxClient, itemID); err != nil {
				u.PrintFatal("Failed to delete file", err)
			}
		}
		u.PrintSuccess("Deleted successfully")
	},
}

func init() {
	deleteCmd.Flags().StringVarP(&deleteFlags.itemID, "id", "i", "", "Delete by item ID instead of path")
	deleteCmd.Flags().BoolVarP(&deleteFlags.recursive, "recursive", "r", false, "Recursively trash a non-empty folder")
	rootCmd.AddCommand(deleteCmd)
}
