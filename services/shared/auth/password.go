// Package auth provides password hashing and verification utilities.
package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// DefaultCost is the bcrypt cost used for password hashing.
// Higher = slower but more secure.  Adjust based on your server's capacity.
const DefaultCost = bcrypt.DefaultCost

// HashPassword hashes a plaintext password using bcrypt.
// Returns the hashed password as a string.
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedBytes), nil
}

// HashPasswordWithCost hashes a password using a specific bcrypt cost.
func HashPasswordWithCost(password string, cost int) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedBytes), nil
}

// VerifyPassword compares a plaintext password against a bcrypt hash.
// Returns nil on match, or an error if the password is incorrect.
func VerifyPassword(hashedPassword, plainPassword string) error {
	if hashedPassword == "" || plainPassword == "" {
		return fmt.Errorf("password and hash must not be empty")
	}
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
}

// IsHashedPassword checks whether a string looks like a bcrypt hash.
func IsHashedPassword(s string) bool {
	// bcrypt hashes start with $2a$, $2b$, or $2y$ and are 60 chars long
	return len(s) == 60 && (s[:4] == "$2a$" || s[:4] == "$2b$" || s[:4] == "$2y$")
}

// ValidatePasswordStrength checks if a password meets minimum requirements.
// Returns an error describing the first requirement that is not met.
func ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	if len(password) > 128 {
		return fmt.Errorf("password must not exceed 128 characters")
	}

	var hasUpper, hasLower, hasDigit bool
	for _, c := range password {
		switch {
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= 'a' && c <= 'z':
			hasLower = true
		case c >= '0' && c <= '9':
			hasDigit = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}

	return nil
}
