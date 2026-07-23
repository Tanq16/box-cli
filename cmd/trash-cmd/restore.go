package trashcmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tanq16/box/cmd/cmdutil"
	"github.com/tanq16/box/internal/api"
	u "github.com/tanq16/box/utils"
)

var restoreFlags struct {
	itemID string
	to     string
}

var RestoreCmd = &cobra.Command{
	Use:   "restore <name>",
	Short: "Restore a trashed file or folder",
	Args:  cobra.MaximumNArgs(1),
	Run: func(c *cobra.Command, args []string) {
		name := ""
		if len(args) > 0 {
			name = args[0]
		}
		if name == "" && restoreFlags.itemID == "" {
			u.PrintFatal("Must specify a name or --id", nil)
		}

		itemID, itemType, err := api.ResolveTrashItem(cmdutil.BoxClient, name, restoreFlags.itemID)
		if err != nil {
			u.PrintFatal("Failed to find trashed item", err)
		}

		parentID := ""
		if restoreFlags.to != "" {
			id, _, err := api.ResolvePath(cmdutil.BoxClient, restoreFlags.to, "folder")
			if err != nil {
				u.PrintFatal("Failed to resolve target folder", err)
			}
			parentID = id
		}

		item, err := api.RestoreItem(cmdutil.BoxClient, itemType, itemID, parentID)
		if err != nil {
			u.PrintFatal("Failed to restore item", err)
		}
		u.PrintSuccess(fmt.Sprintf("Restored %s '%s'", item.Type, item.Name))
	},
}

func init() {
	RestoreCmd.Flags().StringVarP(&restoreFlags.itemID, "id", "i", "", "Restore by item ID instead of name")
	RestoreCmd.Flags().StringVarP(&restoreFlags.to, "to", "t", "", "Target folder path used only if the original parent is gone")
}
