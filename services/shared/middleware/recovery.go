package middleware

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
)

// RecoveryMiddleware recovers from panics and logs the stack trace.
// Returns a clean JSON 500 response to the client.
func RecoveryMiddleware() gin.HandlerFunc {
	logger := log.New(os.Stderr, "[RECOVERY] ", log.LstdFlags|log.Lshortfile)

	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := string(debug.Stack())
				logger.Printf("PANIC: %v\n%s", err, stack)

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error":   "Internal Server Error",
					"code":    "INTERNAL_ERROR",
					"details": "An unexpected error occurred. Please try again later.",
				})
			}
		}()
		c.Next()
	}
}

// RecoveryWithCallback recovers from panics and invokes a callback.
func RecoveryWithCallback(callback func(interface{}, []byte)) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := debug.Stack()
				callback(err, stack)

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "Internal Server Error",
					"code":  "INTERNAL_ERROR",
				})
			}
		}()
		c.Next()
	}
}

// RequestLogger logs request details with structured timing info.
func RequestLogger(logger *log.Logger) gin.HandlerFunc {
	if logger == nil {
		logger = log.New(os.Stdout, "[HTTP] ", log.LstdFlags)
	}

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		logger.Printf("| %3d | %13v | %15s | %-7s %s | %s",
			statusCode,
			latency,
			clientIP,
			method,
			path,
			c.Errors.ByType(gin.ErrorTypePrivate).String(),
		)
	}
}

// RecoveryMiddlewareWithLogger is an alias for RecoveryMiddleware with a custom logger.
func RecoveryMiddlewareWithLogger(logger *log.Logger) gin.HandlerFunc {
	if logger == nil {
		logger = log.New(os.Stderr, "[RECOVERY] ", log.LstdFlags|log.Lshortfile)
	}
	return RecoveryMiddleware()
}

// fmtErrorf is a helper for compatibility.
func fmtErrorf(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}
