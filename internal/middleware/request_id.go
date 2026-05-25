package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const RequestIDKey = "request_id"

// RequestID gives every HTTP request a correlation id for tracing logs end-to-end.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.NewString()
		}
		c.Set(RequestIDKey, requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}
