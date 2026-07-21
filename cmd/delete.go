package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tanq16/box/cmd/cmdutil"
	"github.com/tanq16/box/internal/api"
	u "github.com/tanq16/box/utils"
)

var deleteFlags struct {
	itemID string
}

var deleteCmd = &cobra.Command{
	Use:     "delete <path>",
	Aliases: []string{"rm"},
	Short:   "Delete a file or folder on Box",
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		itemID, itemType := cmdutil.ResolveItem(args, deleteFlags.itemID)

		if itemType == "folder" {
			if err := api.DeleteFolder(boxClient, itemID); err != nil {
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
	rootCmd.AddCommand(deleteCmd)
}
