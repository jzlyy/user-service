package utils

import (
	"context"
	"errors"
	"time"
	"user-service/config"
	"user-service/database"

	"github.com/golang-jwt/jwt/v5"
)

var tokenBlacklistTTL = 24 * time.Hour // 令牌黑名单有效期

func GenerateToken(userID int) (string, error) {
	cfg := config.LoadConfig()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	})

	return token.SignedString([]byte(cfg.JWTSecret))
}

func ParseToken(tokenString string) (int, error) {
	cfg := config.LoadConfig()
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWTSecret), nil
	})

	// 检查令牌是否在黑名单
	if isTokenBlacklisted(tokenString) {
		return 0, errors.New("token is blacklisted")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := int(claims["user_id"].(float64))
		return userID, nil
	}

	return 0, err
}

// 检查令牌是否在黑名单
func isTokenBlacklisted(tokenString string) bool {
	ctx := context.Background()
	key := "jwt_blacklist:" + tokenString
	exists, err := database.RedisClient.Exists(ctx, key).Result()
	return err == nil && exists > 0
}

// BlacklistToken 添加令牌到黑名单
func BlacklistToken(tokenString string) error {
	ctx := context.Background()
	key := "jwt_blacklist:" + tokenString
	return database.RedisClient.Set(ctx, key, "1", tokenBlacklistTTL).Err()
}

func RefreshToken(tokenString string) (string, error) {
	userID, err := ParseToken(tokenString)
	if err != nil {
		return "", err
	}
	return GenerateToken(userID)
}
