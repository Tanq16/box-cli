package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tanq16/box/internal/auth"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Box.com via OAuth",
	RunE: func(cmd *cobra.Command, args []string) error {
		return auth.Login()
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
