package database

import (
	"context"
	"time"
	"user-service/config"

	"github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client

func InitRedis() error {
	cfg := config.LoadConfig()

	RedisClient = redis.NewClient(&redis.Options{
		Addr:         cfg.RedisHost + ":" + cfg.RedisPort,
		Password:     cfg.RedisPassword,
		DB:           cfg.RedisDB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     20, // 连接池大小
		MinIdleConns: 5,  // 最小空闲连接
		IdleTimeout:  5 * time.Minute,
	})

	// 健康检查
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if _, err := RedisClient.Ping(ctx).Result(); err != nil {
		return err
	}
	return nil
}

func CloseRedis() {
	if RedisClient != nil {
		_ = RedisClient.Close()
	}
}
