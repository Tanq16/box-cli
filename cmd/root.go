package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/tanq16/box/cmd/cmdutil"
	collabcmd "github.com/tanq16/box/cmd/collab-cmd"
	sharedlinkcmd "github.com/tanq16/box/cmd/shared-link-cmd"
	synccmd "github.com/tanq16/box/cmd/sync-cmd"
	trashcmd "github.com/tanq16/box/cmd/trash-cmd"
	"github.com/tanq16/box/internal/auth"
	"github.com/tanq16/box/internal/client"
	u "github.com/tanq16/box/utils"
)

var AppVersion = "dev-build"
var debugFlag bool
var forAIFlag bool

var boxClient *client.BoxClient

var rootCmd = &cobra.Command{
	Use:     "box",
	Short:   "CLI tool for Box.com file operations",
	Version: AppVersion,
	CompletionOptions: cobra.CompletionOptions{
		HiddenDefaultCmd: true,
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if cmd.Name() == "login" {
			return
		}
		httpClient, err := auth.GetHTTPClient()
		if err != nil {
			u.PrintFatal("failed to authenticate — run 'box login' first", err)
		}
		boxClient = client.New(httpClient)
		cmdutil.BoxClient = boxClient
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func setupLogs() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.DateTime,
		NoColor:    false,
	}
	log.Logger = zerolog.New(output).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debugFlag {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		u.GlobalDebugFlag = true
	}
	if forAIFlag {
		u.GlobalForAIFlag = true
		zerolog.SetGlobalLevel(zerolog.Disabled)
	}
}

func init() {
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "Enable debug logging")
	rootCmd.PersistentFlags().BoolVar(&forAIFlag, "for-ai", false, "AI-friendly output (plain text, piped input)")
	rootCmd.MarkFlagsMutuallyExclusive("debug", "for-ai")

	cobra.OnInitialize(setupLogs)

	rootCmd.AddCommand(collabcmd.CollabCmd)
	rootCmd.AddCommand(sharedlinkcmd.SharedLinkCmd)
	rootCmd.AddCommand(synccmd.SyncCmd)
	rootCmd.AddCommand(trashcmd.TrashCmd)
	rootCmd.AddCommand(trashcmd.RestoreCmd)
}
