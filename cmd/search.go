package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/tanq16/box/internal/api"
	"github.com/tanq16/box/internal/types"
	"github.com/tanq16/box/internal/utils"
)

var (
	searchType          string
	searchExtensions    string
	searchFolderID      string
	searchCreatedAfter  string
	searchCreatedBefore string
	searchUpdatedAfter  string
	searchUpdatedBefore string
	searchSizeMin       int64
	searchSizeMax       int64
	searchOwner         string
	searchSort          string
	searchLimit         int
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for files and folders on Box (server-side)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := args[0]

		var exts []string
		if searchExtensions != "" {
			exts = strings.Split(searchExtensions, ",")
		}

		opts := types.SearchOptions{
			Query:         query,
			Type:          searchType,
			Extensions:    exts,
			FolderID:      searchFolderID,
			CreatedAfter:  searchCreatedAfter,
			CreatedBefore: searchCreatedBefore,
			UpdatedAfter:  searchUpdatedAfter,
			UpdatedBefore: searchUpdatedBefore,
			SizeMin:       searchSizeMin,
			SizeMax:       searchSizeMax,
			Owner:         searchOwner,
			Sort:          searchSort,
			Limit:         searchLimit,
		}

		results, err := api.Search(boxClient, opts)
		if err != nil {
			return err
		}

		if len(results.Entries) == 0 {
			fmt.Println("No results found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "TYPE\tID\tNAME\tSIZE")
		for _, item := range results.Entries {
			size := "-"
			if item.Size != nil {
				size = utils.FormatSize(*item.Size)
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", item.Type, item.ID, item.Name, size)
		}
		w.Flush()
		fmt.Fprintf(os.Stderr, "\n%d results (of %d total)\n", len(results.Entries), results.TotalCount)
		return nil
	},
}

func init() {
	searchCmd.Flags().StringVar(&searchType, "type", "", "Filter by type (file, folder, web_link)")
	searchCmd.Flags().StringVar(&searchExtensions, "extensions", "", "Comma-separated file extensions")
	searchCmd.Flags().StringVar(&searchFolderID, "folder-id", "", "Search within folder ID")
	searchCmd.Flags().StringVar(&searchCreatedAfter, "created-after", "", "Created after (RFC3339)")
	searchCmd.Flags().StringVar(&searchCreatedBefore, "created-before", "", "Created before (RFC3339)")
	searchCmd.Flags().StringVar(&searchUpdatedAfter, "updated-after", "", "Updated after (RFC3339)")
	searchCmd.Flags().StringVar(&searchUpdatedBefore, "updated-before", "", "Updated before (RFC3339)")
	searchCmd.Flags().Int64Var(&searchSizeMin, "size-min", 0, "Minimum file size in bytes")
	searchCmd.Flags().Int64Var(&searchSizeMax, "size-max", 0, "Maximum file size in bytes")
	searchCmd.Flags().StringVar(&searchOwner, "owner", "", "Owner user ID")
	searchCmd.Flags().StringVar(&searchSort, "sort", "", "Sort by (modified_at or relevance)")
	searchCmd.Flags().IntVar(&searchLimit, "limit", 30, "Maximum number of results")
	rootCmd.AddCommand(searchCmd)
}
