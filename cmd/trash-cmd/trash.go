package trashcmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tanq16/box/cmd/cmdutil"
	"github.com/tanq16/box/internal/api"
	u "github.com/tanq16/box/utils"
)

var TrashCmd = &cobra.Command{
	Use:   "trash",
	Short: "Manage the Box trash",
}

var trashListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List items in the trash",
	Run: func(c *cobra.Command, args []string) {
		items, err := api.ListTrash(cmdutil.BoxClient)
		if err != nil {
			u.PrintFatal("Failed to list trash", err)
		}
		if len(items) == 0 {
			u.PrintInfo("Trash is empty")
			return
		}

		headers := []string{"TYPE", "ID", "NAME", "SIZE", "MODIFIED"}
		var rows [][]string
		for _, item := range items {
			size := "-"
			if item.Type != "folder" {
				size = u.FormatSize(item.Size)
			}
			rows = append(rows, []string{item.Type, item.ID, item.Name, size, item.ModifiedTime})
		}
		u.PrintTable(headers, rows)
	},
}

var trashEmptyFlags struct {
	yes bool
}

var trashEmptyCmd = &cobra.Command{
	Use:   "empty",
	Short: "Permanently delete everything in the trash (cannot be undone)",
	Run: func(c *cobra.Command, args []string) {
		if !trashEmptyFlags.yes && !u.GlobalForAIFlag {
			if !cmdutil.Confirm("Permanently delete ALL trashed items?") {
				u.PrintInfo("Aborted")
				return
			}
		}
		count, err := api.EmptyTrash(cmdutil.BoxClient)
		if err != nil {
			u.PrintFatal("Failed to empty trash", err)
		}
		u.PrintSuccess(fmt.Sprintf("Purged %d item(s) from trash", count))
	},
}

func init() {
	trashEmptyCmd.Flags().BoolVarP(&trashEmptyFlags.yes, "yes", "y", false, "Skip confirmation")
	TrashCmd.AddCommand(trashListCmd)
	TrashCmd.AddCommand(trashEmptyCmd)
}
