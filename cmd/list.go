package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/tanq16/box/internal/api"
	"github.com/tanq16/box/internal/utils"
)

var listFolderID string

var listCmd = &cobra.Command{
	Use:   "list [path]",
	Short: "List contents of a Box folder",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
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
				return err
			}
		}

		folders, files, err := api.ListFolder(boxClient, folderID)
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "TYPE\tID\tNAME\tSIZE\tMODIFIED")
		for _, f := range folders {
			fmt.Fprintf(w, "folder\t%s\t%s\t-\t%s\n", f.ID, f.Name, f.ModifiedTime)
		}
		for _, f := range files {
			fmt.Fprintf(w, "file\t%s\t%s\t%s\t%s\n", f.ID, f.Name, utils.FormatSize(f.Size), f.ModifiedTime)
		}
		w.Flush()
		return nil
	},
}

func init() {
	listCmd.Flags().StringVar(&listFolderID, "id", "", "List by folder ID instead of path")
	rootCmd.AddCommand(listCmd)
}
