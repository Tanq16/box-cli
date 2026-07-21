package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tanq16/box/internal/auth"
	u "github.com/tanq16/box/utils"
)

var loginFlags struct {
	manual bool
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Box.com via OAuth",
	Run: func(cmd *cobra.Command, args []string) {
		config, err := auth.LoadCredentials()
		if err != nil {
			u.PrintFatal("failed to load credentials", err)
		}

		mode := "default"
		if loginFlags.manual {
			mode = "manual"
		}

		token, err := auth.Login(config, mode)
		if err != nil {
			u.PrintFatal("login failed", err)
		}

		_ = token
		u.PrintSuccess("authenticated successfully — token saved")
	},
}

func init() {
	loginCmd.Flags().BoolVar(&loginFlags.manual, "manual", false, "Manually paste authorization code (if browser flow fails)")
	rootCmd.AddCommand(loginCmd)
}
