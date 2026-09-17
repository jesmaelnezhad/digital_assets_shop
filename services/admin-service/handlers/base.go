package handlers

import (
	"database/sql"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// db is the package-level database connection pool.
var db *sql.DB

// InitDB sets the package-level database connection.
func InitDB(database *sql.DB) {
	db = database
}

// GetDB returns the package-level database connection.
func GetDB() *sql.DB {
	return db
}

// AdminAuthMiddleware validates the admin bearer token.
func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		adminToken := os.Getenv("ADMIN_TOKEN")
		if adminToken == "" {
			adminToken = "admin-secret-token-change-in-production"
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Admin authorization required"})
			c.Abort()
			return
		}

		const prefix = "Bearer "
		if len(authHeader) < len(prefix) || authHeader[:len(prefix)] != prefix {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Admin authorization required"})
			c.Abort()
			return
		}

		token := authHeader[len(prefix):]
		if token != adminToken {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid admin token"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// nullTimeToString safely converts a sql.NullTime to a string.
func nullTimeToString(t sql.NullTime) interface{} {
	if t.Valid {
		return t.Time.Format("2006-01-02T15:04:05Z")
	}
	return nil
}

// parseInt parses a string to int, returns 0 on error.
func parseInt(s string) int {
	n := 0
	neg := false
	for i, c := range s {
		if i == 0 && c == '-' {
			neg = true
			continue
		}
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int(c-'0')
	}
	if neg {
		n = -n
	}
	return n
}

// itoa converts an integer to a string.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
