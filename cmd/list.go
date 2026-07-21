package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tanq16/box/internal/api"
	u "github.com/tanq16/box/utils"
)

var listFlags struct {
	folderID string
	filter   string
}

var listCmd = &cobra.Command{
	Use:     "list [path]",
	Aliases: []string{"ls"},
	Short:   "List contents of a Box folder",
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		remotePath := "/"
		if len(args) > 0 {
			remotePath = args[0]
		}

		items, err := api.ListFolderItems(boxClient, listFlags.folderID, remotePath, listFlags.filter)
		if err != nil {
			u.PrintFatal("Failed to list folder", err)
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

func init() {
	listCmd.Flags().StringVarP(&listFlags.folderID, "id", "i", "", "List by folder ID instead of path")
	listCmd.Flags().StringVarP(&listFlags.filter, "filter", "F", "", "Case-insensitive substring filter on item names")
	rootCmd.AddCommand(listCmd)
}
