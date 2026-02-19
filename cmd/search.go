package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tanq16/box/internal/api"
	"github.com/tanq16/box/internal/types"
	u "github.com/tanq16/box/internal/utils"
)

var (
	searchType       string
	searchExtensions string
	searchFolderID   string
	searchCreatedIn  string
	searchUpdatedIn  string
	searchSizeMin    int64
	searchSizeMax    int64
	searchOwner      string
	searchSort       string
	searchLimit      int
)

// parseRelativeTime parses a shorthand like "2h", "3d", "1w", "2M", "1y"
// and returns the RFC3339 timestamp for that duration ago from now.
// Supported units: s (seconds), m (minutes), h (hours), d (days), w (weeks), M (months), y (years).
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
		if searchExtensions != "" {
			exts = strings.Split(searchExtensions, ",")
		}

		createdAfter, err := parseRelativeTime(searchCreatedIn)
		if err != nil {
			u.PrintFatal("Invalid --created-in value", err)
		}

		updatedAfter, err := parseRelativeTime(searchUpdatedIn)
		if err != nil {
			u.PrintFatal("Invalid --updated-in value", err)
		}

		opts := types.SearchOptions{
			Query:        query,
			Type:         searchType,
			Extensions:   exts,
			FolderID:     searchFolderID,
			CreatedAfter: createdAfter,
			UpdatedAfter: updatedAfter,
			SizeMin:      searchSizeMin,
			SizeMax:      searchSizeMax,
			Owner:        searchOwner,
			Sort:         searchSort,
			Limit:        searchLimit,
		}

		results, err := api.Search(boxClient, opts)
		if err != nil {
			u.PrintFatal("Search failed", err)
		}

		if len(results.Entries) == 0 {
			u.PrintInfo("No results found")
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
		u.PrintInfo(fmt.Sprintf("%d results (of %d total)", len(results.Entries), results.TotalCount))
	},
}

func init() {
	searchCmd.Flags().StringVar(&searchType, "type", "", "Filter by type (file, folder, web_link)")
	searchCmd.Flags().StringVar(&searchExtensions, "extensions", "", "Comma-separated file extensions")
	searchCmd.Flags().StringVar(&searchFolderID, "folder-id", "", "Search within folder ID")
	searchCmd.Flags().StringVar(&searchCreatedIn, "created-in", "", "Created within relative time (e.g. 2h, 3d, 1w, 2M, 1y)")
	searchCmd.Flags().StringVar(&searchUpdatedIn, "updated-in", "", "Updated within relative time (e.g. 2h, 3d, 1w, 2M, 1y)")
	searchCmd.Flags().Int64Var(&searchSizeMin, "size-min", 0, "Minimum file size in bytes")
	searchCmd.Flags().Int64Var(&searchSizeMax, "size-max", 0, "Maximum file size in bytes")
	searchCmd.Flags().StringVar(&searchOwner, "owner", "", "Owner user ID")
	searchCmd.Flags().StringVar(&searchSort, "sort", "", "Sort by (modified_at or relevance)")
	searchCmd.Flags().IntVar(&searchLimit, "limit", 30, "Maximum number of results")
	rootCmd.AddCommand(searchCmd)
}
