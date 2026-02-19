package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tanq16/box/internal/api"
	u "github.com/tanq16/box/internal/utils"
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
	Run: func(cmd *cobra.Command, args []string) {
		localDir := args[0]
		remotePath := args[1]

		info, err := os.Stat(localDir)
		if err != nil || !info.IsDir() {
			u.PrintFatal(fmt.Sprintf("'%s' is not a valid directory", localDir), err)
		}

		ignore := parseIgnore(syncIgnore)
		u.PrintInfo(fmt.Sprintf("Syncing push: %s → %s (concurrency: %d)", localDir, remotePath, syncConcurrency))
		if err := api.SyncPush(boxClient, localDir, remotePath, syncConcurrency, ignore); err != nil {
			u.PrintFatal("Sync push failed", err)
		}
		u.PrintSuccess("Sync push complete")
	},
}

var syncPullCmd = &cobra.Command{
	Use:   "pull <remote> <local>",
	Short: "Pull remote changes to local (remote is truth)",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		remotePath := args[0]
		localDir := args[1]

		ignore := parseIgnore(syncIgnore)
		u.PrintInfo(fmt.Sprintf("Syncing pull: %s → %s (concurrency: %d)", remotePath, localDir, syncConcurrency))
		if err := api.SyncPull(boxClient, remotePath, localDir, syncConcurrency, ignore); err != nil {
			u.PrintFatal("Sync pull failed", err)
		}
		u.PrintSuccess("Sync pull complete")
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
