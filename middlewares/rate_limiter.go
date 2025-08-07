package middlewares

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"
	"user-service/database"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"golang.org/x/time/rate"
)

func RateLimiter() gin.HandlerFunc {
	localLimiter := rate.NewLimiter(100, 30) // 本地限流

	return func(c *gin.Context) {
		// 全局分布式限流（按IP）
		ip := c.ClientIP()
		key := "rate_limit:" + ip

		// 使用Redis滑动窗口
		now := time.Now().UnixNano()
		windowSize := time.Minute
		clearBefore := now - windowSize.Nanoseconds()

		ctx := context.Background()
		// 移除时间窗口外的请求
		_, err := database.RedisClient.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(clearBefore, 10)).Result()
		if err != nil {
			log.Printf("Redis ZRemRangeByScore error: %v", err)
		}

		// 获取当前请求数
		count, err := database.RedisClient.ZCard(ctx, key).Result()
		if err != nil {
			log.Printf("Redis ZCard error: %v", err)
			c.Next()
			return
		}

		// 检查是否超过限制
		if count > 50 {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "too many global requests",
			})
			return
		}

		// 添加当前请求
		_, err = database.RedisClient.ZAdd(ctx, key, &redis.Z{
			Score:  float64(now),
			Member: now,
		}).Result()
		if err != nil {
			log.Printf("Redis ZAdd error: %v", err)
		}

		// 设置键过期时间
		database.RedisClient.Expire(ctx, key, windowSize)

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
