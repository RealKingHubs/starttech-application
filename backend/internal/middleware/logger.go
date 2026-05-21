package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// StructuredLogger returns a Gin middleware that provides structured, contextual logging for every request.
func StructuredLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		requestID := uuid.New().String()

		// Create a logger with context for this request
		requestLogger := slog.With(
			slog.String("request_id", requestID),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.String("ip", c.ClientIP()),
			slog.String("user_agent", c.Request.UserAgent()),
		)

		// Set the logger in the context so it can be used by handlers
		c.Set("logger", requestLogger)
		requestLogger.Info("request started")

		// Process the request
		c.Next()

		// Log request completion details
		latency := time.Since(start)
		requestLogger.Info("request completed",
			slog.Int("status", c.Writer.Status()),
			slog.Duration("latency", latency),
		)
	}
}
