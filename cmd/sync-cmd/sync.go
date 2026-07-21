package synccmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"time"

	"github.com/spf13/cobra"
	"github.com/tanq16/box/cmd/cmdutil"
	"github.com/tanq16/box/internal/api"
	u "github.com/tanq16/box/utils"
)

var syncFlags struct {
	concurrency int
	ignore      string
}

var SyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync files between local and Box",
}

var syncPushCmd = &cobra.Command{
	Use:   "push <local> <remote>",
	Short: "Push local changes to Box (local is truth)",
	Args:  cobra.ExactArgs(2),
	Run: func(c *cobra.Command, args []string) {
		localDir := args[0]
		remotePath := args[1]

		info, err := os.Stat(localDir)
		if err != nil || !info.IsDir() {
			u.PrintFatal(fmt.Sprintf("'%s' is not a valid directory", localDir), err)
		}

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()

		ignore := parseIgnore(syncFlags.ignore)

		u.PrintRunning("Scanning files...")
		plan, err := api.PlanPush(ctx, cmdutil.BoxClient, localDir, remotePath, ignore)
		if err != nil {
			u.ClearLines(1)
			u.PrintFatal("Sync push failed", err)
		}
		u.ClearLines(1)

		if plan.Total == 0 {
			u.PrintSuccess("Already in sync")
			return
		}

		u.PrintRunning(fmt.Sprintf("Syncing push: %d to upload, %d to update, %d to delete", plan.Add, plan.Update, plan.Delete))

		progress := &api.SyncProgress{}
		done := make(chan struct{})
		var printed atomic.Bool
		go func() {
			ticker := time.NewTicker(500 * time.Millisecond)
			defer ticker.Stop()
			firstTick := true
			for {
				select {
				case <-done:
					return
				case <-ticker.C:
					if !firstTick {
						u.ClearPreviousLine()
					}
					firstTick = false
					printed.Store(true)
					pct := int(progress.Completed.Load()) * 100 / plan.Total
					u.PrintProgress("Syncing", pct)
				}
			}
		}()

		err = api.ExecPush(ctx, cmdutil.BoxClient, plan, syncFlags.concurrency, progress)
		close(done)
		if printed.Load() {
			u.ClearPreviousLine()
		}
		u.ClearLines(1)

		if err != nil {
			u.PrintFatal("Sync push failed", err)
		}

		errors := int(progress.Errors.Load())
		if errors > 0 {
			u.PrintWarn(fmt.Sprintf("Sync push complete with %d errors (use --debug for details)", errors), nil)
		} else {
			u.PrintSuccess("Sync push complete")
		}
	},
}

var syncPullCmd = &cobra.Command{
	Use:   "pull <remote> <local>",
	Short: "Pull remote changes to local (remote is truth)",
	Args:  cobra.ExactArgs(2),
	Run: func(c *cobra.Command, args []string) {
		remotePath := args[0]
		localDir := args[1]

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()

		ignore := parseIgnore(syncFlags.ignore)

		u.PrintRunning("Scanning files...")
		plan, err := api.PlanPull(ctx, cmdutil.BoxClient, remotePath, localDir, ignore)
		if err != nil {
			u.ClearLines(1)
			u.PrintFatal("Sync pull failed", err)
		}
		u.ClearLines(1)

		if plan.Total == 0 {
			u.PrintSuccess("Already in sync")
			return
		}

		u.PrintRunning(fmt.Sprintf("Syncing pull: %d to download, %d to update, %d to delete", plan.Add, plan.Update, plan.Delete))

		progress := &api.SyncProgress{}
		done := make(chan struct{})
		var printed atomic.Bool
		go func() {
			ticker := time.NewTicker(500 * time.Millisecond)
			defer ticker.Stop()
			firstTick := true
			for {
				select {
				case <-done:
					return
				case <-ticker.C:
					if !firstTick {
						u.ClearPreviousLine()
					}
					firstTick = false
					printed.Store(true)
					pct := int(progress.Completed.Load()) * 100 / plan.Total
					u.PrintProgress("Syncing", pct)
				}
			}
		}()

		err = api.ExecPull(ctx, cmdutil.BoxClient, plan, syncFlags.concurrency, progress)
		close(done)
		if printed.Load() {
			u.ClearPreviousLine()
		}
		u.ClearLines(1)

		if err != nil {
			u.PrintFatal("Sync pull failed", err)
		}

		errors := int(progress.Errors.Load())
		if errors > 0 {
			u.PrintWarn(fmt.Sprintf("Sync pull complete with %d errors (use --debug for details)", errors), nil)
		} else {
			u.PrintSuccess("Sync pull complete")
		}
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
	SyncCmd.PersistentFlags().IntVarP(&syncFlags.concurrency, "concurrency", "c", 4, "Number of concurrent operations")
	SyncCmd.PersistentFlags().StringVarP(&syncFlags.ignore, "ignore", "i", "", "Comma-separated list of names to ignore")
	SyncCmd.AddCommand(syncPushCmd)
	SyncCmd.AddCommand(syncPullCmd)
}
