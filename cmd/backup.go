package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"syne-cli/internal/db"

	"github.com/spf13/cobra"
)

var (
	format            string
	file              string
	compress          bool
	cleanFirst        bool
	dataOnly          bool
	singleTransaction bool
	schemaOnly        bool
	disableTriggers   bool
)

func init() {
	// Backup command
	backupCmd := &cobra.Command{
		Use:   "backup",
		Short: "Backup a PostgreSQL database",
		Long: `Backup a PostgreSQL database using pg_dump.
Supported formats:
- custom (c): PostgreSQL custom format
- plain (p): Plain SQL script
- directory (d): Directory format
- tar (t): Tar format

Data only flag:
- data-only (D): Backup only data (not schema)`,
		RunE: runBackup,
	}

	backupCmd.Flags().StringVarP(&format, "format", "F", "custom", "Backup format (custom, plain, directory, tar)")
	backupCmd.Flags().StringVarP(&file, "file", "f", "", "Output file name")
	backupCmd.Flags().BoolVarP(&compress, "compress", "Z", true, "Compress backup (for custom and tar formats)")

	// Add PostgreSQL connection flags
	backupCmd.Flags().String("host", "localhost", "PostgreSQL host")
	backupCmd.Flags().String("port", "5432", "PostgreSQL port")
	backupCmd.Flags().String("db-user", "", "PostgreSQL user")
	backupCmd.Flags().String("db-password", "", "PostgreSQL password")
	backupCmd.Flags().String("db-name", "", "PostgreSQL database name")
	backupCmd.Flags().String("ssl-mode", "disable", "PostgreSQL SSL mode")

	// Mark required flags
	backupCmd.MarkFlagRequired("db-user")
	backupCmd.MarkFlagRequired("db-password")
	backupCmd.MarkFlagRequired("db-name")

	// Data only flag
	backupCmd.Flags().BoolVarP(&dataOnly, "data-only", "D", false, "Backup only data (not schema)")
	backupCmd.Flags().BoolVarP(&schemaOnly, "schema-only", "S", false, "Backup only schema (not data)")
	// Restore command
	restoreCmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore a PostgreSQL database",
		Long: `Restore a PostgreSQL database from a backup file.
Automatically detects format from file extension:
- .sql: Plain SQL script
- .dump: Custom format
- .dir: Directory format
- .tar: Tar format`,
		RunE: runRestore,
	}

	restoreCmd.Flags().StringVarP(&file, "file", "f", "", "Input backup file")
	restoreCmd.Flags().BoolVarP(&cleanFirst, "clean", "c", false, "Clean (drop) database objects before recreating")
	restoreCmd.Flags().BoolVarP(&singleTransaction, "single-transaction", "s", false, "Use a single transaction for the restore")

	// Add PostgreSQL connection flags
	restoreCmd.Flags().String("host", "localhost", "PostgreSQL host")
	restoreCmd.Flags().String("port", "5432", "PostgreSQL port")
	restoreCmd.Flags().String("db-user", "", "PostgreSQL user")
	restoreCmd.Flags().String("db-password", "", "PostgreSQL password")
	restoreCmd.Flags().String("db-name", "", "PostgreSQL database name")
	restoreCmd.Flags().String("ssl-mode", "disable", "PostgreSQL SSL mode")

	// skip ownership flag
	restoreCmd.Flags().Bool("skip-ownership", false, "Skip ownership of database objects")

	// restore command
	restoreCmd.Flags().BoolVarP(&schemaOnly, "schema-only", "S", false, "Restore only schema (not data)")
	restoreCmd.Flags().BoolVarP(&dataOnly, "data-only", "D", false, "Backup only data (not schema)")
	restoreCmd.Flags().BoolVarP(&disableTriggers, "disable-triggers", "T", false, "Disable triggers during restore")

	// Mark required flags
	restoreCmd.MarkFlagRequired("db-user")
	restoreCmd.MarkFlagRequired("db-password")
	restoreCmd.MarkFlagRequired("db-name")
	// Add commands to root
	rootCmd.AddCommand(backupCmd)
	rootCmd.AddCommand(restoreCmd)
}

func getFormatFlag(format string) string {
	if len(format) == 1 {
		return format
	}

	formatMap := map[string]string{
		"custom":    "c",
		"plain":     "p",
		"directory": "d",
		"tar":       "t",
	}

	if flag, ok := formatMap[strings.ToLower(format)]; ok {
		return flag
	}
	return "c" // default to custom format
}

func decideFileFormat(format string) string {
	formatMap := map[string]string{
		"p": ".sql",
		"c": ".dump",
		"d": ".dir",
		"t": ".tar",
	}

	if flag, ok := formatMap[format]; ok {
		return flag
	}
	return ".dump" // default to custom format
}

func detectFormat(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	formatMap := map[string]string{
		".sql":  "p",
		".dump": "c",
		".dir":  "d",
		".tar":  "t",
	}

	if flag, ok := formatMap[ext]; ok {
		return flag
	}
	return "c" // default to custom format
}

func runBackup(cmd *cobra.Command, cmdArgs []string) error {
	// Get PostgreSQL connection details from flags
	host, _ := cmd.Flags().GetString("host")
	port, _ := cmd.Flags().GetString("port")
	dbUser, _ := cmd.Flags().GetString("db-user")
	dbPass, _ := cmd.Flags().GetString("db-password")
	dbName, _ := cmd.Flags().GetString("db-name")
	sslMode, _ := cmd.Flags().GetString("ssl-mode")
	dataOnly, _ := cmd.Flags().GetBool("data-only")
	skipOwnership, _ := cmd.Flags().GetBool("skip-ownership")

	config := db.PostgresConfig{
		Host:     host,
		Port:     port,
		User:     dbUser,
		Password: dbPass,
		DBName:   dbName,
		SSLMode:  sslMode,
	}

	var backupFormat string = getFormatFlag(format)
	var backupFile string = decideFileFormat(backupFormat)

	var cmdArgs1 []string
	// Build pg_dump command
	cmdArgs1 = []string{
		"-h", config.Host,
		"-p", config.Port,
		"-U", config.User,
		"-d", config.DBName,
		"-F", backupFormat,
	}

	if skipOwnership {
		cmdArgs1 = append(cmdArgs1, "--no-owner")
	}

	if dataOnly {
		cmdArgs1 = append(cmdArgs1, "--data-only")
	}

	if compress && format != "t" && format != "tar" {
		cmdArgs1 = append(cmdArgs1, "-Z", "9")
	}

	if file == "" {
		file = fmt.Sprintf("%s_%s%s", config.DBName, "backup", backupFile)
	}

	cmdArgs1 = append(cmdArgs1, "-f", file)

	var cmd1 *exec.Cmd = exec.Command("pg_dump", cmdArgs1...)
	cmd1.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", config.Password))

	// Set up pipes for real-time output
	stdout, err := cmd1.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error creating stdout pipe: %v", err)
	}
	stderr, err := cmd1.StderrPipe()
	if err != nil {
		return fmt.Errorf("error creating stderr pipe: %v", err)
	}

	// Start the command
	if err := cmd1.Start(); err != nil {
		return fmt.Errorf("error starting pg_dump: %v", err)
	}

	// Create scanner for real-time output
	go func() {
		scanner := bufio.NewScanner(io.MultiReader(stdout, stderr))
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()

	if err := cmd1.Wait(); err != nil {
		return fmt.Errorf("backup failed: %v", err)
	}

	fmt.Printf("Successfully backed up database %s to %s\n", config.DBName, file)
	return nil
}

func runRestore(cmd *cobra.Command, cmdArgs []string) error {
	if file == "" {
		return fmt.Errorf("input file is required")
	}

	// Get PostgreSQL connection details from flags
	host, _ := cmd.Flags().GetString("host")
	port, _ := cmd.Flags().GetString("port")
	dbUser, _ := cmd.Flags().GetString("db-user")
	dbPass, _ := cmd.Flags().GetString("db-password")
	dbName, _ := cmd.Flags().GetString("db-name")
	sslMode, _ := cmd.Flags().GetString("ssl-mode")
	skipOwnership, _ := cmd.Flags().GetBool("skip-ownership")
	disableTriggers, _ := cmd.Flags().GetBool("disable-triggers")

	singleTransaction, _ := cmd.Flags().GetBool("single-transaction")
	schemaOnly, _ := cmd.Flags().GetBool("schema-only")
	config := db.PostgresConfig{
		Host:     host,
		Port:     port,
		User:     dbUser,
		Password: dbPass,
		DBName:   dbName,
		SSLMode:  sslMode,
	}

	dataOnly, _ := cmd.Flags().GetBool("data-only")

	format = detectFormat(file)

	var cmd1 *exec.Cmd
	var cmdArgs1 []string

	// For plain SQL format, use psql
	if format == "p" {
		cmdArgs1 = []string{
			"-h", config.Host,
			"-p", config.Port,
			"-U", config.User,
			"-d", config.DBName,
			"-f", file,
			"-v",
		}

		if singleTransaction {
			cmdArgs1 = append(cmdArgs1, "--single-transaction")
		}

		if skipOwnership {
			cmdArgs1 = append(cmdArgs1, "--no-owner")
		}

		if schemaOnly {
			cmdArgs1 = append(cmdArgs1, "--schema-only")
		}

		if dataOnly {
			cmdArgs1 = append(cmdArgs1, "--data-only")
		}

		if disableTriggers {
			cmdArgs1 = append(cmdArgs1, "--disable-triggers")
		}

		cmd1 = exec.Command("psql", cmdArgs1...)
		cmd1.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", config.Password))

		// Set up pipes for real-time output
		stdout, err := cmd1.StdoutPipe()
		if err != nil {
			return fmt.Errorf("error creating stdout pipe: %v", err)
		}
		stderr, err := cmd1.StderrPipe()
		if err != nil {
			return fmt.Errorf("error creating stderr pipe: %v", err)
		}

		// Start the command
		if err := cmd1.Start(); err != nil {
			return fmt.Errorf("error starting psql: %v", err)
		}

		// Create scanner for real-time output
		go func() {
			scanner := bufio.NewScanner(io.MultiReader(stdout, stderr))
			for scanner.Scan() {
				fmt.Println(scanner.Text())
			}
		}()

		if err := cmd1.Wait(); err != nil {
			return fmt.Errorf("restore failed: %v", err)
		}
	} else {
		// For other formats, use pg_restore
		cmdArgs1 = []string{
			"-h", config.Host,
			"-p", config.Port,
			"-U", config.User,
			"-d", config.DBName,
			"-v", // Add verbose flag to show progress
		}

		if cleanFirst {
			cmdArgs1 = append(cmdArgs1, "--clean")
		}

		if singleTransaction {
			cmdArgs1 = append(cmdArgs1, "--single-transaction")
		}

		if skipOwnership {
			cmdArgs1 = append(cmdArgs1, "--no-owner")
		}

		if schemaOnly {
			cmdArgs1 = append(cmdArgs1, "--schema-only")
		}

		if dataOnly {
			cmdArgs1 = append(cmdArgs1, "--data-only")
		}

		if disableTriggers {
			cmdArgs1 = append(cmdArgs1, "--disable-triggers")
		}

		cmdArgs1 = append(cmdArgs1, file)

		cmd1 = exec.Command("pg_restore", cmdArgs1...)
		cmd1.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", config.Password))

		// Set up pipes for real-time output
		stdout, err := cmd1.StdoutPipe()
		if err != nil {
			return fmt.Errorf("error creating stdout pipe: %v", err)
		}
		stderr, err := cmd1.StderrPipe()
		if err != nil {
			return fmt.Errorf("error creating stderr pipe: %v", err)
		}

		// Start the command
		if err := cmd1.Start(); err != nil {
			return fmt.Errorf("error starting pg_restore: %v", err)
		}

		// Create scanner for real-time output
		go func() {
			scanner := bufio.NewScanner(io.MultiReader(stdout, stderr))
			for scanner.Scan() {
				fmt.Println(scanner.Text())
			}
		}()

		if err := cmd1.Wait(); err != nil {
			return fmt.Errorf("restore failed: %v", err)
		}
	}

	fmt.Printf("Successfully restored database %s from %s\n", config.DBName, file)
	return nil
}
