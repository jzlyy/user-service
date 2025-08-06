package middlewares

import (
	"context"
	"net/http"
	"time"
	"user-service/database"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func RateLimiter() gin.HandlerFunc {
	localLimiter := rate.NewLimiter(100, 30) // 本地限流

	return func(c *gin.Context) {
		// 全局分布式限流（按IP）
		ip := c.ClientIP()
		key := "rate_limit:" + ip

		ctx := context.Background()
		count, err := database.RedisClient.Incr(ctx, key).Result()
		if err == nil {
			if count == 1 {
				// 设置过期时间
				database.RedisClient.Expire(ctx, key, time.Minute)
			}

			if count > 50 { // 全局限制50次/分钟
				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
					"error": "too many global requests",
				})
				return
			}
		}

		// 本地限流
		if !localLimiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "too many local requests",
			})
			return
		}
		c.Next()
	}
}
