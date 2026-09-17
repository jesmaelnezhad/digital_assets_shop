package middleware

import (
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// LoggingMiddleware logs HTTP requests with timing and status information.
// It writes to stderr by default, or a custom logger can be passed.
func LoggingMiddleware() gin.HandlerFunc {
	logger := log.New(os.Stderr, "[HTTP] ", log.LstdFlags)
	return LoggingMiddlewareWithLogger(logger)
}

// LoggingMiddlewareWithLogger logs requests using a custom logger.
func LoggingMiddlewareWithLogger(logger *log.Logger) gin.HandlerFunc {
	if logger == nil {
		logger = log.New(os.Stderr, "[HTTP] ", log.LstdFlags)
	}

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
			c.Header("X-Request-ID", requestID)
			c.Set("request_id", requestID)
		}

		c.Next()

		latency := time.Since(start)
		_ = c.ClientIP() // for future use
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		logger.Printf("[%s] %3d | %13v | %-7s %s | %s",
			requestID,
			statusCode,
			latency,
			method,
			path,
			c.Errors.ByType(gin.ErrorTypePrivate).String(),
		)
	}
}

// generateRequestID creates a simple unique request identifier.
// In production use a UUID library.
func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
