package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tanq16/box/internal/api"
	"github.com/tanq16/box/internal/types"
	u "github.com/tanq16/box/utils"
)

var searchFlags struct {
	itemType   string
	extensions string
	folderID   string
	createdIn  string
	updatedIn  string
	sizeMin    int64
	sizeMax    int64
	owner      string
	sort       string
	limit      int
}

func parseRelativeTime(s string) (string, error) {
	if s == "" {
		return "", nil
	}

	s = strings.TrimSpace(s)
	if len(s) < 2 {
		return "", fmt.Errorf("invalid relative time: %q (need number + unit, e.g. 2h, 3d, 1M)", s)
	}

	unit := s[len(s)-1:]
	numStr := s[:len(s)-1]
	n, err := strconv.Atoi(numStr)
	if err != nil {
		return "", fmt.Errorf("invalid relative time: %q (bad number %q)", s, numStr)
	}

	now := time.Now().UTC()
	var t time.Time
	switch unit {
	case "s":
		t = now.Add(-time.Duration(n) * time.Second)
	case "m":
		t = now.Add(-time.Duration(n) * time.Minute)
	case "h":
		t = now.Add(-time.Duration(n) * time.Hour)
	case "d":
		t = now.AddDate(0, 0, -n)
	case "w":
		t = now.AddDate(0, 0, -n*7)
	case "M":
		t = now.AddDate(0, -n, 0)
	case "y":
		t = now.AddDate(-n, 0, 0)
	default:
		return "", fmt.Errorf("unknown time unit %q (use s, m, h, d, w, M, or y)", unit)
	}

	return t.Format(time.RFC3339), nil
}

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for files and folders on Box (server-side)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]

		var exts []string
		if searchFlags.extensions != "" {
			exts = strings.Split(searchFlags.extensions, ",")
		}

		createdAfter, err := parseRelativeTime(searchFlags.createdIn)
		if err != nil {
			u.PrintFatal("cmd","Invalid --created-in value", err)
		}

		updatedAfter, err := parseRelativeTime(searchFlags.updatedIn)
		if err != nil {
			u.PrintFatal("cmd","Invalid --updated-in value", err)
		}

		opts := types.SearchOptions{
			Query:        query,
			Type:         searchFlags.itemType,
			Extensions:   exts,
			FolderID:     searchFlags.folderID,
			CreatedAfter: createdAfter,
			UpdatedAfter: updatedAfter,
			SizeMin:      searchFlags.sizeMin,
			SizeMax:      searchFlags.sizeMax,
			Owner:        searchFlags.owner,
			Sort:         searchFlags.sort,
			Limit:        searchFlags.limit,
		}

		results, err := api.Search(boxClient, opts)
		if err != nil {
			u.PrintFatal("cmd","Search failed", err)
		}

		if len(results.Entries) == 0 {
			u.PrintInfo("cmd","No results found")
			return
		}

		headers := []string{"TYPE", "ID", "NAME", "SIZE"}
		var rows [][]string
		for _, item := range results.Entries {
			size := "-"
			if item.Size != nil {
				size = u.FormatSize(*item.Size)
			}
			rows = append(rows, []string{item.Type, item.ID, item.Name, size})
		}
		u.PrintTable(headers, rows)
	},
}

func init() {
	searchCmd.Flags().StringVar(&searchFlags.itemType, "type", "", "Filter by type (file, folder, web_link)")
	searchCmd.Flags().StringVar(&searchFlags.extensions, "extensions", "", "Comma-separated file extensions")
	searchCmd.Flags().StringVar(&searchFlags.folderID, "folder-id", "", "Search within folder ID")
	searchCmd.Flags().StringVar(&searchFlags.createdIn, "created-in", "", "Created within relative time (e.g. 2h, 3d, 1w, 2M, 1y)")
	searchCmd.Flags().StringVar(&searchFlags.updatedIn, "updated-in", "", "Updated within relative time (e.g. 2h, 3d, 1w, 2M, 1y)")
	searchCmd.Flags().Int64Var(&searchFlags.sizeMin, "size-min", 0, "Minimum file size in bytes")
	searchCmd.Flags().Int64Var(&searchFlags.sizeMax, "size-max", 0, "Maximum file size in bytes")
	searchCmd.Flags().StringVar(&searchFlags.owner, "owner", "", "Owner user ID")
	searchCmd.Flags().StringVar(&searchFlags.sort, "sort", "", "Sort by (modified_at or relevance)")
	searchCmd.Flags().IntVar(&searchFlags.limit, "limit", 30, "Maximum number of results")
	rootCmd.AddCommand(searchCmd)
}
