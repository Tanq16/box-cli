package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tanq16/box/internal/api"
)

var syncConcurrency int
var syncIgnore string

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync files between local and Box",
}

var syncPushCmd = &cobra.Command{
	Use:   "push <local> <remote>",
	Short: "Push local changes to Box (local is truth)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		localDir := args[0]
		remotePath := args[1]

		info, err := os.Stat(localDir)
		if err != nil || !info.IsDir() {
			return fmt.Errorf("'%s' is not a valid directory", localDir)
		}

		ignore := parseIgnore(syncIgnore)
		fmt.Fprintf(os.Stderr, "Syncing push: %s → %s (concurrency: %d)\n", localDir, remotePath, syncConcurrency)
		return api.SyncPush(boxClient, localDir, remotePath, syncConcurrency, ignore)
	},
}

var syncPullCmd = &cobra.Command{
	Use:   "pull <remote> <local>",
	Short: "Pull remote changes to local (remote is truth)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		remotePath := args[0]
		localDir := args[1]

		ignore := parseIgnore(syncIgnore)
		fmt.Fprintf(os.Stderr, "Syncing pull: %s → %s (concurrency: %d)\n", remotePath, localDir, syncConcurrency)
		return api.SyncPull(boxClient, remotePath, localDir, syncConcurrency, ignore)
	},
}

func parseIgnore(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func init() {
	syncCmd.PersistentFlags().IntVar(&syncConcurrency, "concurrency", 4, "Number of concurrent operations")
	syncCmd.PersistentFlags().StringVar(&syncIgnore, "ignore", "", "Comma-separated list of names to ignore")
	syncCmd.AddCommand(syncPushCmd)
	syncCmd.AddCommand(syncPullCmd)
	rootCmd.AddCommand(syncCmd)
}
