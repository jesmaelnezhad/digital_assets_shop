package middleware

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pawradise/shared/auth"
)

// JwtAuthMiddleware validates JWT tokens from the Authorization header or cookie.
// It extracts the user_id and email from claims and stores them in context.
// If no token is present, the request continues (for optional-auth routes).
func JwtAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try Bearer header first, then cookie
		tokenString := extractToken(c)
		if tokenString == "" {
			c.Next()
			return
		}

		secret := getJwtSecret()
		claims, err := auth.ValidateJWT(tokenString, secret)
		if err != nil || claims == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// Store user info and continue
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Set("token_hash", auth.HashToken(tokenString))
		c.Next()
	}
}

// extractToken tries Bearer header first, then cookie
func extractToken(c *gin.Context) string {
	if auth := c.GetHeader("Authorization"); strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	if cookie, err := c.Cookie("pawradise_session"); err == nil {
		return cookie
	}
	return ""
}

// AdminAuthMiddleware restricts access to admin-only endpoints.
// Checks for either a JWT with role=admin or an ADMIN_TOKEN bearer.
func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization required"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Check admin token first
		adminToken := getAdminToken()
		if tokenString == adminToken {
			c.Set("role", "admin")
			c.Next()
			return
		}

		// Check JWT for admin role
		claims, err := auth.ValidateJWT(tokenString, getJwtSecret())
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		if claims.Role == "admin" {
			c.Set("user_id", claims.UserID)
			c.Set("role", "admin")
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
	}
}

// GetUserIDFromContext extracts the authenticated user_id from context.
// Returns 0 and false if no user_id is present.
func GetUserIDFromContext(c *gin.Context) (int, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	switch v := userID.(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	default:
		return 0, false
	}
}

// RequireAuth is a strict middleware variant: aborts if no valid token.
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := GetUserIDFromContext(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			return
		}
		_ = userID
		c.Next()
	}
}

// TimeoutMiddleware adds a request timeout.
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		done := make(chan struct{})
		go func() {
			defer func() {
				if r := recover(); r != nil {
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
						"error": "internal server error",
					})
				}
			}()
			c.Next()
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(timeout):
			c.AbortWithStatusJSON(http.StatusGatewayTimeout, gin.H{
				"error": "request timeout",
			})
		}
	}
}

func getJwtSecret() string {
	if value := os.Getenv("JWT_SECRET"); value != "" {
		return value
	}
	return "your-secret-key-change-in-production"
}

func getAdminToken() string {
	if value := os.Getenv("ADMIN_TOKEN"); value != "" {
		return value
	}
	return "admin-secret-token-change-in-production"
}
