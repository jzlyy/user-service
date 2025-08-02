package utils

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-redis/redis/v8"
	"strconv"
	"time"
	"user-service/database"
)

const (
	UserCacheTTL = 30 * time.Minute // 用户数据缓存时间
)

func SetUserCache(userID int, user interface{}) error {
	ctx := context.Background()
	key := userCacheKey(userID)

	jsonData, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return database.RedisClient.Set(ctx, key, jsonData, UserCacheTTL).Err()
}

func GetUserCache(userID int, dest interface{}) (bool, error) {
	ctx := context.Background()
	key := userCacheKey(userID)

	data, err := database.RedisClient.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return false, nil // 缓存未命中
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
