package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func RateLimiter(rds *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		key := "rate_limit:" + ctx.ClientIP()
		pipe := rds.TxPipeline()
		// Increment the request count
		incr := pipe.Incr(ctx, key)
		// Set the expiration time if this is the first request
		pipe.Expire(ctx, key, window)

		_, err := pipe.Exec(ctx)
		if err != nil {
			ctx.AbortWithStatusJSON(500, gin.H{"error": "Internal server error"})
			return
		}

		if incr.Val() > int64(limit) {
			ctx.AbortWithStatusJSON(429, gin.H{"error": "Rate limit exceeded"})
			return
		}

		ctx.Next()
	}
}
