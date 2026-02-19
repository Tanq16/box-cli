package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/tanq16/box/internal/auth"
	"github.com/tanq16/box/internal/client"
	"github.com/tanq16/box/internal/utils"
)

var AppVersion = "dev-build"
var debugFlag bool

var boxClient *client.BoxClient

var rootCmd = &cobra.Command{
	Use:     "box",
	Short:   "CLI tool for Box.com file operations",
	Version: AppVersion,
	CompletionOptions: cobra.CompletionOptions{
		HiddenDefaultCmd: true,
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Skip auth for login command
		if cmd.Name() == "login" {
			return
		}
		httpClient, err := auth.GetClient()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		boxClient = client.New(httpClient)
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
		utils.GlobalDebugFlag = true
	}
}

func init() {
	// Hide default help and completion commands
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	// Global debug flag
	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "Enable debug logging")

	// Initialize logging on startup
	cobra.OnInitialize(setupLogs)
}
