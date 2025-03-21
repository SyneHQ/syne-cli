package cmd

import (
	"fmt"
	"syne-cli/internal/auth"

	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to the CLI",
	Long:  `Login to the CLI using your username and password.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")

		if username == "" || password == "" {
			return fmt.Errorf("username and password are required")
		}

		if err := auth.Authenticate(username, password); err != nil {
			return fmt.Errorf("authentication failed: %v", err)
		}

		fmt.Println("Successfully logged in!")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
