package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tanq16/box/internal/auth"
	"github.com/tanq16/box/internal/client"
)

var boxClient *client.BoxClient

var rootCmd = &cobra.Command{
	Use:   "box",
	Short: "CLI tool for Box.com file operations",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip auth for login command
		if cmd.Name() == "login" {
			return nil
		}
		httpClient, err := auth.GetClient()
		if err != nil {
			return err
		}
		boxClient = client.New(httpClient)
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
