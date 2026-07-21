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
	dryRun      bool
	delete      bool
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

		if syncFlags.dryRun {
			printPlan(plan, false)
			return
		}

		if plan.HasDeletes() && !confirmDeletes(plan) {
			return
		}

		deleteCount := 0
		if syncFlags.delete {
			deleteCount = plan.Delete
		}
		u.PrintRunning(fmt.Sprintf("Syncing push: %d to upload, %d to update, %d to delete", plan.Add, plan.Update, deleteCount))

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

		err = api.ExecPush(ctx, cmdutil.BoxClient, plan, syncFlags.concurrency, progress, syncFlags.delete)
		close(done)
		if printed.Load() {
			u.ClearPreviousLine()
		}
		u.ClearLines(1)

		if err != nil {
			u.PrintFatal("Sync push failed", err)
		}

		reportResult("Sync push", progress)
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

		if syncFlags.dryRun {
			printPlan(plan, true)
			return
		}

		if plan.HasDeletes() && !confirmDeletes(plan) {
			return
		}

		deleteCount := 0
		if syncFlags.delete {
			deleteCount = plan.Delete
		}
		u.PrintRunning(fmt.Sprintf("Syncing pull: %d to download, %d to update, %d to delete", plan.Add, plan.Update, deleteCount))

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

		err = api.ExecPull(ctx, cmdutil.BoxClient, plan, syncFlags.concurrency, progress, syncFlags.delete)
		close(done)
		if printed.Load() {
			u.ClearPreviousLine()
		}
		u.ClearLines(1)

		if err != nil {
			u.PrintFatal("Sync pull failed", err)
		}

		reportResult("Sync pull", progress)
	},
}

func printPlan(plan *api.SyncPlan, download bool) {
	verb := "upload"
	if download {
		verb = "download"
	}
	u.PrintInfo(fmt.Sprintf("Dry run: %d to %s, %d to update, %d to delete, %d folder change(s)", plan.Add, verb, plan.Update, plan.Delete, plan.Folders))
	for _, op := range plan.Ops {
		line := planLabel(op.Action, download) + op.Path
		if op.Action == api.SyncDelete || op.Action == api.SyncDeleteFolder {
			u.PrintIndentedWarn(line, nil)
		} else {
			u.PrintIndentedSuccess(line)
		}
	}
	if plan.HasDeletes() && !syncFlags.delete {
		u.PrintWarn("Deletes shown are previewed only; pass --delete to apply them", nil)
	}
}

func planLabel(action api.SyncAction, download bool) string {
	switch action {
	case api.SyncAdd:
		if download {
			return "download "
		}
		return "upload "
	case api.SyncUpdate:
		return "update "
	case api.SyncDelete:
		return "delete "
	case api.SyncCreateFolder:
		return "mkdir "
	case api.SyncDeleteFolder:
		return "rmdir "
	}
	return ""
}

func confirmDeletes(plan *api.SyncPlan) bool {
	if !syncFlags.delete {
		u.PrintWarn(fmt.Sprintf("Skipping %d delete(s); pass --delete to remove items missing from the source", plan.DeleteTotal()), nil)
		return true
	}
	if u.GlobalForAIFlag {
		return true
	}
	if cmdutil.Confirm(fmt.Sprintf("Delete %d item(s) missing from the source?", plan.DeleteTotal())) {
		return true
	}
	u.PrintInfo("Aborted")
	return false
}

func reportResult(label string, progress *api.SyncProgress) {
	errors := int(progress.Errors.Load())
	if errors == 0 {
		u.PrintSuccess(label + " complete")
		return
	}
	u.PrintWarn(fmt.Sprintf("%s complete with %d error(s)", label, errors), nil)
	for _, f := range progress.Failures {
		u.PrintIndentedError(f.Item, f.Err)
	}
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
	SyncCmd.PersistentFlags().BoolVar(&syncFlags.dryRun, "dry-run", false, "Show planned changes without executing them")
	SyncCmd.PersistentFlags().BoolVar(&syncFlags.delete, "delete", false, "Delete items missing from the source (destructive)")
	SyncCmd.AddCommand(syncPushCmd)
	SyncCmd.AddCommand(syncPullCmd)
}
