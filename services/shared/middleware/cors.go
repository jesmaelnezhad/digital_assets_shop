package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware configures Cross-Origin Resource Sharing headers.
// It reads allowed origins from the CORS_ORIGINS env var (comma-separated).
// If CORS_ORIGINS is empty, it auto-detects based on ENV_NAME:
//   - staging → allows the server's own domain (same-origin via ingress)
//   - production → allows the production domain
// For cross-origin requests from different domains, set CORS_ORIGINS explicitly.
func CORSMiddleware() gin.HandlerFunc {
	return CORS()
}

// CORS is a shorthand for CORSMiddleware (called CORS() in service code).
func CORS() gin.HandlerFunc {
	envName := os.Getenv("ENV_NAME")
	allowedOrigins := getAllowedOrigins(envName)

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// Check if origin is in the allowed list
		if isOriginAllowed(origin, allowedOrigins) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Accept, Origin, Cache-Control, X-Requested-With, X-Admin-Token")
			c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")
			c.Header("Access-Control-Max-Age", "86400")
		}

		// Handle preflight
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func getAllowedOrigins(envName string) []string {
	env := os.Getenv("CORS_ORIGINS")
	if env != "" {
		parts := strings.Split(env, ",")
		var origins []string
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				origins = append(origins, trimmed)
			}
		}
		return origins
	}

	// Auto-detect based on environment
	stagingHost := "server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir"
	productionHost := "pawradise.ir" // placeholder - replace with real domain

	switch envName {
	case "staging":
		return []string{"https://" + stagingHost, "http://" + stagingHost}
	case "production":
		return []string{"https://" + productionHost, "http://" + productionHost}
	default:
		// Default permissive origins for development
		return []string{"http://localhost:3000", "http://localhost:5173", "http://localhost:8080"}
	}
}

func isOriginAllowed(origin string, allowed []string) bool {
	if origin == "" {
		return true // same-origin request
	}
	for _, o := range allowed {
		if o == "*" || o == origin {
			return true
		}
	}
	return false
}
