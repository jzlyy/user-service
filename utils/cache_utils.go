package utils

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-redis/redis/v8"
	"log"
	"strconv"
	"time"
	"user-service/config"
	"user-service/database"
	"user-service/services"
)

// CacheLockTimeout 获取缓存锁的超时时间
const CacheLockTimeout = 2 * time.Second

// 函数获取TTL
func getUserCacheTTL() time.Duration {
	cfg := config.LoadConfig()
	return time.Duration(cfg.CacheTTL) * time.Minute
}

func SetUserCache(userID int, user interface{}) error {
	ctx := context.Background()
	key := userCacheKey(userID)
	ttl := getUserCacheTTL() // 动态获取

	jsonData, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return database.RedisClient.Set(ctx, key, jsonData, ttl).Err()
}

func GetUserCache(userID int, dest interface{}) (bool, error) {
	ctx := context.Background()
	key := userCacheKey(userID)

	data, err := database.RedisClient.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		// 尝试获取缓存锁
		lockKey := "lock:" + key
		if AcquireLock(lockKey, CacheLockTimeout) {
			defer ReleaseLock(lockKey)

			// 从数据库加载 - 使用services包
			user, err := services.GetUserByID(userID)
			if err != nil {
				return false, err
			}

			// 设置缓存
			if err := SetUserCache(userID, user); err != nil {
				log.Printf("Failed to set user cache: %v", err)
			}

			// 转换数据
			jsonData, _ := json.Marshal(user)
			return true, json.Unmarshal(jsonData, dest)
		} else {
			// 等待其他请求完成缓存
			time.Sleep(500 * time.Millisecond)
			return GetUserCache(userID, dest) // 重试
		}

	} else if err != nil {
		return false, err
	}

	return true, json.Unmarshal(data, dest)
}

func InvalidateUserCache(userID int) error {
	ctx := context.Background()
	key := userCacheKey(userID)
	return database.RedisClient.Del(ctx, key).Err()
}

func userCacheKey(userID int) string {
	return "user:" + strconv.Itoa(userID)
}
