package auth

import (
	"crypto/subtle"
	"fmt"
	"log/slog"
	"strings"

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

// CheckPasswordHash compares a password with what is stored for an SMPP credential: a bcrypt
// hash, or (for credentials created before hashing) the password itself, compared in constant
// time. Use IsBcryptHash to tell which one matched, so legacy values can be upgraded.
func CheckPasswordHash(password, hash string) bool {
	if IsBcryptHash(hash) {
		return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
	}
	return subtle.ConstantTimeCompare([]byte(password), []byte(hash)) == 1
}

// IsBcryptHash reports whether a stored value is a bcrypt hash rather than a legacy plain value.
func IsBcryptHash(stored string) bool {
	return strings.HasPrefix(stored, "$2a$") || strings.HasPrefix(stored, "$2b$") || strings.HasPrefix(stored, "$2y$")
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
