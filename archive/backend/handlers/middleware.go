package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strings"

	"backend/database"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// hashToken returns the SHA256 hex of a token (never store raw tokens).
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// isTokenRevoked reports whether the token hash is in the blocklist.
func isTokenRevoked(tokenHash string) bool {
	if database.DB == nil {
		return false
	}
	var one int
	if err := database.DB.QueryRow(`SELECT 1 FROM invalidated_tokens WHERE token_hash = $1`, tokenHash).Scan(&one); err != nil {
		return false
	}
	return true
}

func JwtAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.Next()
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		secret := JwtSecret()
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}
		if isTokenRevoked(hashToken(tokenString)) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token revoked"})
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}
		if userID, ok := claims["user_id"].(float64); ok {
			c.Set("user_id", int(userID))
		}
		c.Next()
	}
}

func JwtSecret() string {
	if value := os.Getenv("JWT_SECRET"); value != "" {
		return value
	}
	return "your-secret-key-change-in-production"
}
