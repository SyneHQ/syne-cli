package cmd

import (
	"fmt"
	"syne-cli/internal/auth"
	"syne-cli/internal/db"

	"github.com/spf13/cobra"
)

var copyCmd = &cobra.Command{
	Use:   "copy [folder] [table]",
	Short: "Copy files from a folder to PostgreSQL table",
	Long:  `Copy all files from the specified folder to a PostgreSQL table using the COPY command.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		// First authenticate
		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")

		if username == "" || password == "" {
			return fmt.Errorf("username and password are required")
		}

		if err := auth.Authenticate(username, password); err != nil {
			return fmt.Errorf("authentication failed: %v", err)
		}

		// Get PostgreSQL connection details from flags
		host, _ := cmd.Flags().GetString("host")
		port, _ := cmd.Flags().GetString("port")
		dbUser, _ := cmd.Flags().GetString("db-user")
		dbPass, _ := cmd.Flags().GetString("db-password")
		dbName, _ := cmd.Flags().GetString("db-name")
		sslMode, _ := cmd.Flags().GetString("ssl-mode")

		// Create database connection
		config := db.PostgresConfig{
			Host:     host,
			Port:     port,
			User:     dbUser,
			Password: dbPass,
			DBName:   dbName,
			SSLMode:  sslMode,
		}

		conn, err := db.Connect(config)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %v", err)
		}
		defer conn.Close()

		folderPath := args[0]
		tableName := args[1]

		// Perform the copy operation
		if err := db.CopyFolder(conn, folderPath, tableName); err != nil {
			return fmt.Errorf("copy operation failed: %v", err)
		}

		fmt.Println("Copy operation completed successfully!")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(copyCmd)

	// Add PostgreSQL connection flags
	copyCmd.Flags().String("host", "localhost", "PostgreSQL host")
	copyCmd.Flags().String("port", "5432", "PostgreSQL port")
	copyCmd.Flags().String("db-user", "", "PostgreSQL user")
	copyCmd.Flags().String("db-password", "", "PostgreSQL password")
	copyCmd.Flags().String("db-name", "", "PostgreSQL database name")
	copyCmd.Flags().String("ssl-mode", "disable", "PostgreSQL SSL mode")

	// Mark required flags
	copyCmd.MarkFlagRequired("db-user")
	copyCmd.MarkFlagRequired("db-password")
	copyCmd.MarkFlagRequired("db-name")
}
