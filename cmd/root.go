package cmd

import (
	"github.com/spf13/cobra"
)

var (
	version string
	rootCmd = &cobra.Command{
		Use:     "syne-cli",
		Short:   "A CLI tool for PostgreSQL operations",
		Long:    `A CLI tool that provides PostgreSQL operations and requires authentication.`,
		Version: version,
	}
)

func SetVersion(v string) {
	version = v
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringP("username", "u", "", "Username for authentication")
	rootCmd.PersistentFlags().StringP("password", "p", "", "Password for authentication")
}
