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
				u.PrintFatal("cmd","Failed to get folder info", err)
			}
			u.PrintGeneric(fmt.Sprintf("Type:     %s", info.Type))
			u.PrintGeneric(fmt.Sprintf("ID:       %s", info.ID))
			u.PrintGeneric(fmt.Sprintf("Name:     %s", info.Name))
			if info.Size != nil {
				u.PrintGeneric(fmt.Sprintf("Size:     %s", u.FormatSize(*info.Size)))
			}
			if info.ModifiedAt != nil {
				if t, err := time.Parse(time.RFC3339, *info.ModifiedAt); err == nil {
					u.PrintGeneric(fmt.Sprintf("Modified: %s", t.Format("2006-01-02 15:04:05")))
				}
			}
			if info.Parent != nil {
				u.PrintGeneric(fmt.Sprintf("Parent:   %s (ID: %s)", info.Parent.Type, info.Parent.ID))
			}
			if info.ItemCollection != nil {
				u.PrintGeneric(fmt.Sprintf("Items:    %d", info.ItemCollection.TotalCount))
			}
		} else {
			info, err := api.GetFileInfo(boxClient, itemID)
			if err != nil {
				u.PrintFatal("cmd","Failed to get file info", err)
			}
			u.PrintGeneric(fmt.Sprintf("Type:     %s", info.Type))
			u.PrintGeneric(fmt.Sprintf("ID:       %s", info.ID))
			u.PrintGeneric(fmt.Sprintf("Name:     %s", info.Name))
			if info.Size != nil {
				u.PrintGeneric(fmt.Sprintf("Size:     %s", u.FormatSize(*info.Size)))
			}
			if info.SHA1 != "" {
				u.PrintGeneric(fmt.Sprintf("SHA1:     %s", info.SHA1))
			}
			if info.ModifiedAt != nil {
				if t, err := time.Parse(time.RFC3339, *info.ModifiedAt); err == nil {
					u.PrintGeneric(fmt.Sprintf("Modified: %s", t.Format("2006-01-02 15:04:05")))
				}
			}
			if info.Parent != nil {
				u.PrintGeneric(fmt.Sprintf("Parent:   %s (ID: %s)", info.Parent.Type, info.Parent.ID))
			}
		}
	},
}

func init() {
	infoCmd.Flags().StringVar(&infoFlags.itemID, "id", "", "Get info by item ID instead of path")
	rootCmd.AddCommand(infoCmd)
}
