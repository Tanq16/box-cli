package cmd

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/tanq16/box/internal/api"
	u "github.com/tanq16/box/internal/utils"
)

var listFolderID string
var listFilter string

var listCmd = &cobra.Command{
	Use:   "list [path]",
	Short: "List contents of a Box folder",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var folderID string
		var err error

		if listFolderID != "" {
			folderID = listFolderID
		} else {
			remotePath := "/"
			if len(args) > 0 {
				remotePath = args[0]
			}
			folderID, _, err = api.ResolvePath(boxClient, remotePath, "folder")
			if err != nil {
				u.PrintFatal("Failed to resolve path", err)
			}
		}

		folders, files, err := api.ListFolder(boxClient, folderID)
		if err != nil {
			u.PrintFatal("Failed to list folder", err)
		}

		filter := strings.ToLower(listFilter)

		headers := []string{"TYPE", "ID", "NAME", "SIZE", "MODIFIED"}
		var rows [][]string
		for _, f := range folders {
			if filter != "" && !strings.Contains(strings.ToLower(f.Name), filter) {
				continue
			}
			rows = append(rows, []string{"folder", f.ID, f.Name, "-", f.ModifiedTime})
		}
		for _, f := range files {
			if filter != "" && !strings.Contains(strings.ToLower(f.Name), filter) {
				continue
			}
			rows = append(rows, []string{"file", f.ID, f.Name, u.FormatSize(f.Size), f.ModifiedTime})
		}
		u.PrintTable(headers, rows)
	},
}

func init() {
	listCmd.Flags().StringVar(&listFolderID, "id", "", "List by folder ID instead of path")
	listCmd.Flags().StringVarP(&listFilter, "filter", "F", "", "Case-insensitive substring filter on item names")
	rootCmd.AddCommand(listCmd)
}
