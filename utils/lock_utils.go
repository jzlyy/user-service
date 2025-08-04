package utils

import (
	"context"
	"time"
	"user-service/database"
)

func AcquireLock(key string, timeout time.Duration) bool {
	ctx := context.Background()
	return database.RedisClient.SetNX(ctx, "lock:"+key, "1", timeout).Val()
}

func ReleaseLock(key string) {
	ctx := context.Background()
	database.RedisClient.Del(ctx, "lock:"+key)
}
