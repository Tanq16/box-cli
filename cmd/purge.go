package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tanq16/box/cmd/cmdutil"
	"github.com/tanq16/box/internal/api"
	u "github.com/tanq16/box/utils"
)

var purgeFlags struct {
	itemID string
	yes    bool
}

var purgeCmd = &cobra.Command{
	Use:   "purge <name>",
	Short: "Permanently delete a trashed file or folder (cannot be undone)",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := ""
		if len(args) > 0 {
			name = args[0]
		}
		if name == "" && purgeFlags.itemID == "" {
			u.PrintFatal("Must specify a name or --id", nil)
		}

		itemID, itemType, err := api.ResolveTrashItem(boxClient, name, purgeFlags.itemID)
		if err != nil {
			u.PrintFatal("Failed to find trashed item", err)
		}

		label := name
		if label == "" {
			label = itemID
		}
		if !purgeFlags.yes && !u.GlobalForAIFlag {
			if !cmdutil.Confirm(fmt.Sprintf("Permanently delete %s '%s'?", itemType, label)) {
				u.PrintInfo("Aborted")
				return
			}
		}

		if itemType == "folder" {
			if err := api.PurgeFolder(boxClient, itemID); err != nil {
				u.PrintFatal("Failed to purge folder", err)
			}
		} else {
			if err := api.PurgeFile(boxClient, itemID); err != nil {
				u.PrintFatal("Failed to purge file", err)
			}
		}
		u.PrintSuccess("Purged permanently")
	},
}

func init() {
	purgeCmd.Flags().StringVarP(&purgeFlags.itemID, "id", "i", "", "Purge by item ID instead of name")
	purgeCmd.Flags().BoolVarP(&purgeFlags.yes, "yes", "y", false, "Skip confirmation")
	rootCmd.AddCommand(purgeCmd)
}
