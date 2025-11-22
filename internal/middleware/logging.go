package middleware

import (
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func SlogMiddleware() gin.HandlerFunc {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	return func(c *gin.Context) {
		start := time.Now()
		reqID := uuid.New().String()
		c.Writer.Header().Set("X-Request-ID", reqID)
		requestLogger := logger.With(slog.String("req_id", reqID))

		requestLogger.Info("request received",
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.String("query", c.Request.URL.RawQuery),
			slog.String("ip", c.ClientIP()),
			slog.String("user-agent", c.Request.UserAgent()),
		)

		c.Next()

		requestLogger.Info("request processed",
			slog.Int("status", c.Writer.Status()),
			slog.Duration("latency", time.Since(start)),
			slog.String("ip", c.ClientIP()),
			slog.String("user-agent", c.Request.UserAgent()),
		)
	}
}
