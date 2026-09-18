package middleware

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/store4bots/shared/auth"
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
		c.Set("role", auth.NormalizeRole(claims.Role))
		c.Set("tabs", claims.Tabs)
		c.Set("token_hash", auth.HashToken(tokenString))
		c.Next()
	}
}

// extractToken tries Bearer header first, then cookie
func extractToken(c *gin.Context) string {
	if auth := c.GetHeader("Authorization"); strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	if cookie, err := c.Cookie("store4bots_session"); err == nil {
		return cookie
	}
	return ""
}

// AdminAuthMiddleware restricts access to staff/admin desk endpoints.
// Static ADMIN_TOKEN bearer is a full-admin API bypass (e2e / automation).
// Staff and admins authenticate with a session JWT (cookie or Bearer).
// Changing roles/tabs additionally requires X-Admin-Token = ADMIN_TOKEN.
func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		adminToken := getAdminToken()
		bearer := ""
		if h := c.GetHeader("Authorization"); strings.HasPrefix(h, "Bearer ") {
			bearer = strings.TrimPrefix(h, "Bearer ")
		}
		if bearer != "" && bearer == adminToken {
			c.Set("role", "admin")
			c.Set("operator", true)
			c.Set("tabs", "*")
			c.Next()
			return
		}

		tokenString := extractToken(c)
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization required"})
			return
		}

		claims, err := auth.ValidateJWT(tokenString, getJwtSecret())
		if err != nil || claims == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		role := auth.NormalizeRole(claims.Role)
		if role != "admin" && role != "staff" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Staff or admin access required"})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", role)
		c.Set("tabs", claims.Tabs)

		operator := false
		if role == "admin" {
			op := c.GetHeader("X-Admin-Token")
			if op == adminToken {
				operator = true
			}
		}
		c.Set("operator", operator)

		if PrivilegedAccessPath(c) {
			if role != "admin" || !operator {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Operator token required"})
				return
			}
			c.Next()
			return
		}

		if role == "admin" {
			c.Next()
			return
		}

		tab := TabForRequest(c)
		if !StaffHasTab(claims.Tabs, tab) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "This section is not assigned to your account"})
			return
		}
		c.Next()
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
