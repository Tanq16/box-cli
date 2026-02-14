package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/tanq16/box/internal/api"
)

var indexSearchType string
var indexSearchPath string

var indexCmd = &cobra.Command{
	Use:   "index [path]",
	Short: "Build a local index of Box contents",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		rootPath := "/"
		if len(args) > 0 {
			rootPath = args[0]
		}
		return api.GenerateIndex(boxClient, rootPath)
	},
}

var indexSearchCmd = &cobra.Command{
	Use:   "search <regex>",
	Short: "Search the local index with a regex pattern",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pattern := args[0]
		results, err := api.SearchIndex(pattern, indexSearchType, indexSearchPath)
		if err != nil {
			return err
		}
		if len(results) == 0 {
			fmt.Println("No matches found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "TYPE\tID\tNAME\tPATH")
		for _, item := range results {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", item.Type, item.ID, item.Name, item.Path)
		}
		w.Flush()
		return nil
	},
}

func init() {
	indexSearchCmd.Flags().StringVar(&indexSearchType, "type", "", "Filter by type (file or folder)")
	indexSearchCmd.Flags().StringVar(&indexSearchPath, "path", "", "Filter by path prefix")
	indexCmd.AddCommand(indexSearchCmd)
	rootCmd.AddCommand(indexCmd)
}
