package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tanq16/box/internal/auth"
	u "github.com/tanq16/box/internal/utils"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Box.com via OAuth",
	Run: func(cmd *cobra.Command, args []string) {
		if err := auth.Login(); err != nil {
			u.PrintFatal("cmd","Authentication failed", err)
		}
		u.PrintSuccess("cmd","Authentication complete")
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
