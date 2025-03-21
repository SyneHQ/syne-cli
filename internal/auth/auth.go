package auth

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
)

func init() {
	godotenv.Load()
}

// Authenticate verifies the username and password
func Authenticate(username, password string) error {
	// For demonstration, we'll use environment variables to store credentials
	// In a real application, you would use a database
	validUsername := os.Getenv("SYNE_CLI_USERNAME")
	hashedPassword := os.Getenv("SYNE_CLI_PASSWORD")

	if username != validUsername {
		return ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return ErrInvalidCredentials
	}

	return nil
}

// HashPassword creates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}
