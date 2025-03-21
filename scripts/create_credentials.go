package main

import (
	"fmt"
	"log"
	"os"
	"syne-cli/internal/auth"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatal("Usage: go run scripts/create_credentials.go <username> <password>")
	}

	username := os.Args[1]
	password := os.Args[2]

	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		log.Fatalf("Error hashing password: %v", err)
	}

	envContent := fmt.Sprintf(`SYNE_CLI_USERNAME=%s
SYNE_CLI_PASSWORD=%s
`, username, hashedPassword)

	if err := os.WriteFile(".env", []byte(envContent), 0644); err != nil {
		log.Fatalf("Error writing .env file: %v", err)
	}

	fmt.Println("Credentials have been saved to .env file")
}
