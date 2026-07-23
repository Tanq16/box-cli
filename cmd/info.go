package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tanq16/box/cmd/cmdutil"
	"github.com/tanq16/box/internal/api"
	u "github.com/tanq16/box/utils"
)

var infoFlags struct {
	itemID string
}

var infoCmd = &cobra.Command{
	Use:   "info <path>",
	Short: "Show metadata for a file or folder on Box",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		itemID, itemType := cmdutil.ResolveItem(args, infoFlags.itemID)

		if itemType == "folder" {
			info, err := api.GetFolderInfo(boxClient, itemID)
			if err != nil {
				u.PrintFatal("Failed to get folder info", err)
			}
			headers := []string{"FIELD", "VALUE"}
			rows := [][]string{
				{"Type", info.Type},
				{"ID", info.ID},
				{"Name", info.Name},
			}
			if info.Size != nil {
				rows = append(rows, []string{"Size", u.FormatSize(*info.Size)})
			}
			if info.ModifiedAt != nil {
				if t, err := time.Parse(time.RFC3339, *info.ModifiedAt); err == nil {
					rows = append(rows, []string{"Modified", t.Format("2006-01-02 15:04:05")})
				}
			}
			if info.Parent != nil {
				rows = append(rows, []string{"Parent", fmt.Sprintf("%s (ID: %s)", info.Parent.Type, info.Parent.ID)})
			}
			if info.ItemCollection != nil {
				rows = append(rows, []string{"Items", fmt.Sprintf("%d", info.ItemCollection.TotalCount)})
			}
			u.PrintTable(headers, rows)
		} else {
			info, err := api.GetFileInfo(boxClient, itemID)
			if err != nil {
				u.PrintFatal("Failed to get file info", err)
			}
			headers := []string{"FIELD", "VALUE"}
			rows := [][]string{
				{"Type", info.Type},
				{"ID", info.ID},
				{"Name", info.Name},
			}
			if info.Size != nil {
				rows = append(rows, []string{"Size", u.FormatSize(*info.Size)})
			}
			if info.SHA1 != "" {
				rows = append(rows, []string{"SHA1", info.SHA1})
			}
			if info.ModifiedAt != nil {
				if t, err := time.Parse(time.RFC3339, *info.ModifiedAt); err == nil {
					rows = append(rows, []string{"Modified", t.Format("2006-01-02 15:04:05")})
				}
			}
			if info.Parent != nil {
				rows = append(rows, []string{"Parent", fmt.Sprintf("%s (ID: %s)", info.Parent.Type, info.Parent.ID)})
			}
			u.PrintTable(headers, rows)
		}
	},
}

func init() {
	infoCmd.Flags().StringVarP(&infoFlags.itemID, "id", "i", "", "Get info by item ID instead of path")
	rootCmd.AddCommand(infoCmd)
}
