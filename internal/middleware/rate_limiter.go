package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimiter is a small Redis fixed-window limiter. It is enough for portfolio learning;
// production systems usually use Lua scripts or a dedicated gateway for stricter atomicity.
func RateLimiter(redisClient *redis.Client, prefix string, limit int64, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := prefix + ":" + c.ClientIP()
		ctx := c.Request.Context()

		count, err := redisClient.Incr(ctx, key).Result()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"success": false, "message": "rate limiter unavailable"})
			return
		}
		if count == 1 {
			_ = redisClient.Expire(ctx, key, window).Err()
		}
		if count > limit {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"success": false, "message": "too many requests"})
			return
		}

		c.Header("X-RateLimit-Limit", formatInt(limit))
		c.Header("X-RateLimit-Remaining", formatInt(max(limit-count, 0)))
		c.Next()
	}
}

func formatInt(value int64) string {
	return strconv.FormatInt(value, 10)
}
