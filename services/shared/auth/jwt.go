// Package auth provides JWT generation, validation, and token revocation
// utilities shared across all Pawradise microservices.
package auth

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenClaims holds the standard JWT claims used by Pawradise.
type TokenClaims struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

// ValidatableToken extends TokenClaims with JWT validation methods.
type ValidatableToken struct {
	TokenClaims
	jwt.RegisteredClaims
}

// TokenLookupFunc queries the database for a revoked-token hash.
// Returns true if the token has been revoked.
type TokenLookupFunc func(hash string) (bool, error)

// defaultTokenLookup checks the invalidated_tokens table in PostgreSQL.
// Replace this function to use a different storage backend.
var defaultTokenLookup TokenLookupFunc = func(hash string) (bool, error) {
	db := getRevocationDB()
	if db == nil {
		return false, nil // no db → nothing is revoked
	}
	var one int
	err := db.QueryRow("SELECT 1 FROM invalidated_tokens WHERE token_hash = $1", hash).Scan(&one)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// revocationDB is set via SetRevocationDB or lazily from the database package.
var revocationDB *sql.DB

// SetRevocationDB sets the database used for token-revocation lookups.
func SetRevocationDB(db *sql.DB) {
	revocationDB = db
}

func getRevocationDB() *sql.DB {
	return revocationDB
}

// GenerateJWT creates a signed JWT for the given user credentials.
// Default expiry is 24 hours.
func GenerateJWT(userID int, email string, role ...string) (string, error) {
	r := "user"
	if len(role) > 0 && role[0] != "" {
		r = role[0]
	}

	claims := ValidatableToken{
		TokenClaims: TokenClaims{
			UserID: userID,
			Email:  email,
			Role:   r,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "pawradise",
			Subject:   fmt.Sprintf("%d", userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(GetJwtSecret()))
}

// ValidateJWT parses and verifies a JWT string.  It returns the ValidatableToken
// on success or an error if the token is malformed, expired, or has an invalid
// signature.
func ValidateJWT(tokenString, secret string) (*ValidatableToken, error) {
	token, err := jwt.ParseWithClaims(tokenString, &ValidatableToken{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	if claims, ok := token.Claims.(*ValidatableToken); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token claims")
}

// ValidateJWTWithDefaultSecret validates a token using the configured secret.
func ValidateJWTWithDefaultSecret(tokenString string) (*ValidatableToken, error) {
	return ValidateJWT(tokenString, GetJwtSecret())
}

// HashToken returns the SHA-256 hex digest of a raw token string.
// Use this to store token hashes (never store raw tokens).
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// IsTokenRevoked checks whether the given token hash appears in the
// revocation list.  If no database is configured it returns false.
func IsTokenRevoked(tokenHash string) bool {
	if tokenHash == "" {
		return false
	}
	revoked, err := defaultTokenLookup(tokenHash)
	if err != nil {
		// Log in production; here we fail open (assume not revoked) to avoid
		// locking users out when the DB is briefly unavailable.
		return false
	}
	return revoked
}

// IsTokenRevokedStrict is like IsTokenRevoked but returns an error when the
// lookup fails instead of failing open.
func IsTokenRevokedStrict(tokenHash string) (bool, error) {
	if tokenHash == "" {
		return false, nil
	}
	return defaultTokenLookup(tokenHash)
}

// RevokeToken inserts a token hash into the invalidation table with an expiry
// matching the token's own exp claim (if parseable), or 24h from now as a
// fallback.
func RevokeToken(tokenString string) error {
	hash := HashToken(tokenString)
	db := getRevocationDB()
	if db == nil {
		return fmt.Errorf("revocation database not configured")
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	if token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{}); err == nil {
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if exp, ok := claims["exp"].(float64); ok {
				expiresAt = time.Unix(int64(exp), 0)
			}
		}
	}

	_, err := db.Exec(
		`INSERT INTO invalidated_tokens (token_hash, expires_at) VALUES ($1, $2)
		 ON CONFLICT (token_hash) DO NOTHING`,
		hash, expiresAt,
	)
	return err
}

// CleanupExpiredRevocations removes expired rows from the invalidated_tokens
// table.  Call this periodically (e.g., from a background job).
func CleanupExpiredRevocations() error {
	db := getRevocationDB()
	if db == nil {
		return fmt.Errorf("revocation database not configured")
	}
	_, err := db.Exec("DELETE FROM invalidated_tokens WHERE expires_at < NOW()")
	return err
}

// GetJwtSecret returns the JWT signing secret from the environment.
func GetJwtSecret() string {
	if value := os.Getenv("JWT_SECRET"); value != "" {
		return value
	}
	return "your-secret-key-change-in-production"
}

// ExtractBearerToken strips the "Bearer " prefix from an Authorization header.
func ExtractBearerToken(header string) (string, bool) {
	const prefix = "Bearer "
	if strings.HasPrefix(header, prefix) {
		return strings.TrimPrefix(header, prefix), true
	}
	return "", false
}
