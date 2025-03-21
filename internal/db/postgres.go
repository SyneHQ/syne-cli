package db

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/lib/pq"
)

// PostgresConfig holds the connection details
type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// Connect establishes a connection to PostgreSQL
func Connect(config PostgresConfig) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode,
	)

	return sql.Open("postgres", connStr)
}

func tableExists(db *sql.DB, tableName string) (bool, error) {
	var exists bool
	err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM pg_tables WHERE tablename = $1)", tableName).Scan(&exists)
	return exists, err
}

// readHeadersAndTypes reads the headers and determines column types from the first two rows
func readHeadersAndTypes(filePath string) ([]string, []string, error) {
	csvFile, err := os.Open(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("error opening file %s: %v", filePath, err)
	}
	defer csvFile.Close()

	reader := csv.NewReader(csvFile)

	// Read headers
	headers, err := reader.Read()
	if err != nil {
		return nil, nil, fmt.Errorf("error reading headers: %v", err)
	}

	// Read first data row to determine types
	firstRow, err := reader.Read()
	if err != nil {
		return nil, nil, fmt.Errorf("error reading first data row: %v", err)
	}

	// Determine PostgreSQL types based on the data
	types := make([]string, len(firstRow))
	for i, value := range firstRow {
		types[i] = inferPostgresType(value)
		fmt.Printf("Column %d: %s\n", i, types[i])
	}

	return headers, types, nil
}

// inferPostgresType tries to determine the PostgreSQL type from a string value
func inferPostgresType(value string) string {
	// Try to parse as smallint
	if match, _ := regexp.MatchString(`^-?\d+$`, value); match {
		if len(value) <= 5 {
			return "SMALLINT"
		} else if len(value) <= 8 {
			return "INTEGER"
		} else if len(value) <= 19 {
			return "BIGINT"
		} else {
			return "TEXT"
		}
	}

	// Try to parse as decimal or numeric, ensuring it has a decimal point
	if match, _ := regexp.MatchString(`^-?\d+\.\d+$`, value); match {
		if len(value) <= 10 {
			return "NUMERIC"
		} else {
			return "TEXT"
		}
	}

	// Try to parse as date
	if match, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}$`, value); match {
		return "DATE"
	}

	// Try to parse as timestamp
	if match, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}`, value); match {
		return "TIMESTAMP"
	}

	// Try to parse as boolean
	value = strings.ToLower(value)
	if value == "true" || value == "false" || value == "t" || value == "f" || value == "yes" || value == "no" {
		return "BOOLEAN"
	}

	// Default to TEXT
	return "TEXT"
}

// CopyFolder copies files from the specified folder to PostgreSQL
func CopyFolder(db *sql.DB, folderPath, tableName string) error {
	// Ensure folder exists
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		return fmt.Errorf("folder does not exist: %s", folderPath)
	}

	// Get list of files in the folder
	files, err := os.ReadDir(folderPath)
	if err != nil {
		return fmt.Errorf("error reading folder: %v", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue // Skip directories
		}

		filePath := filepath.Join(folderPath, file.Name())

		// first check if table exists
		exists, err := tableExists(db, tableName)
		if err != nil {
			return fmt.Errorf("error checking if table exists: %v", err)
		}

		if !exists {
			// Read headers and infer types
			headers, types, err := readHeadersAndTypes(filePath)
			if err != nil {
				return fmt.Errorf("error analyzing CSV structure: %v", err)
			}

			// Create column definitions for CREATE TABLE
			columns := make([]string, len(headers))
			for i := range headers {
				columns[i] = fmt.Sprintf("%s %s", headers[i], types[i])
			}

			// Create the table
			createTableCmd := fmt.Sprintf("CREATE TABLE %s (%s);",
				tableName,
				strings.Join(columns, ", "))
			fmt.Printf("Creating table with command: %s\n", createTableCmd)
			_, err = db.Exec(createTableCmd)
			if err != nil {
				return fmt.Errorf("error creating table: %v", err)
			}
		}

		fmt.Printf("Processing file: %s\n", file.Name())

		// Open the file for reading
		csvFile, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("error opening file %s: %v", filePath, err)
		}
		defer csvFile.Close()

		// Create a CSV reader
		reader := csv.NewReader(csvFile)

		// Get headers
		headers, err := reader.Read()
		if err != nil {
			return fmt.Errorf("error reading CSV headers: %v", err)
		}

		// Start a transaction
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("error starting transaction: %v", err)
		}
		defer tx.Rollback()

		// Create the CopyIn statement
		stmt, err := tx.Prepare(pq.CopyIn(tableName, headers...))
		if err != nil {
			return fmt.Errorf("error preparing copy statement: %v", err)
		}
		defer stmt.Close()

		// Read and copy data
		var rowCount int
		for {
			record, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("error reading CSV row: %v", err)
			}

			// Convert record values to interface{}
			values := make([]interface{}, len(record))
			for i, v := range record {
				values[i] = v
			}

			// Execute the copy statement
			if _, err := stmt.Exec(values...); err != nil {
				return fmt.Errorf("error copying row: %v", err)
			}
			rowCount++
		}

		// Flush the copy statement
		if _, err := stmt.Exec(); err != nil {
			return fmt.Errorf("error flushing copy statement: %v", err)
		}

		// Commit the transaction
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("error committing transaction: %v", err)
		}

		fmt.Printf("Successfully copied %d rows from %s\n", rowCount, file.Name())
	}

	return nil
}
