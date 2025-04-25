package auth

import (
	"errors"
	"fmt"
	"log/slog"

	"golang.org/x/crypto/bcrypt"
)

const (
	bcryptCostFactor = 12
)

// HashPassword generates a bcrypt hash for the given password.
// Use this when creating or updating user/credential passwords before storing in DB.
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCostFactor)
	if err != nil {
		slog.Error("Failed to generate bcrypt hash for password", slog.Any("error", err))
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedBytes), nil
}

// CheckPasswordHash compares a plaintext password with a stored bcrypt hash.
// Use this during SMPP bind authentication.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			// Log unexpected errors during comparison
			slog.Warn("Error comparing password hash", slog.Any("error", err))
		}
		return false // Passwords don't match or error occurred
	}
	return true // Passwords match
}

// HashAPIKey generates a bcrypt hash for the given API key secret.
// Use this when creating or updating SP API keys before storing the hash in DB.
func HashAPIKey(apiKeySecret string) (string, error) {
	// API Keys can often be longer than passwords, bcrypt handles this fine.
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(apiKeySecret), bcryptCostFactor)
	if err != nil {
		slog.Error("Failed to generate bcrypt hash for API key", slog.Any("error", err))
		return "", fmt.Errorf("failed to hash api key: %w", err)
	}
	return string(hashedBytes), nil
}

// CheckAPIKey compares a plaintext API key secret with a stored bcrypt hash.
// Use this during HTTP Authentication middleware.
func CheckAPIKey(apiKeySecret, hash string) bool {
	// Reusing the same bcrypt comparison logic
	return CheckPasswordHash(apiKeySecret, hash)
}
