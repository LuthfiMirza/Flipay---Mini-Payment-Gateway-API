package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RequestLogger writes one structured JSON log per request.
func RequestLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		logger.Info("http request",
			zap.String("request_id", c.GetString(RequestIDKey)),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Int("size", c.Writer.Size()),
			zap.String("client_ip", c.ClientIP()),
			zap.Duration("latency", time.Since(start)),
		)
	}
}
