package utils

import (
	"context"
	"time"
	"user-service/database"
)

func AcquireLock(key string, timeout time.Duration) bool {
	ctx := context.Background()
	lockKey := "lock:" + key
	
	result, err := database.RedisClient.SetNX(ctx, lockKey, "1", timeout).Result()
	if err != nil || !result {
		return false
	}

	// 启动锁续期协程
	go func() {
		ticker := time.NewTicker(timeout / 2)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				renewed, err := database.RedisClient.Expire(ctx, lockKey, timeout).Result()
				if err != nil || !renewed {
					return
				}
			}
		}
	}()

	return true
}

func ReleaseLock(key string) {
	ctx := context.Background()
	database.RedisClient.Del(ctx, "lock:"+key)
}
