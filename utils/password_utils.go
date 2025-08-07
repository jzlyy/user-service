package utils

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"user-service/database"

	"golang.org/x/crypto/bcrypt"
)

const MaxPasswordHistory = 5 // 保留的历史密码数量

func ValidatePasswordStrength(password string) bool {
	// 检查密码是否已泄露
	if isPasswordCompromised(password) {
		return false
	}
	if len(password) < 8 {
		return false
	}

	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case regexp.MustCompile(`[^a-zA-Z0-9]`).MatchString(string(char)):
			hasSpecial = true
		}
	}

	// 要求至少包含大写字母、小写字母、数字和特殊字符中的三种
	criteriaMet := 0
	if hasUpper {
		criteriaMet++
	}
	if hasLower {
		criteriaMet++
	}
	if hasNumber {
		criteriaMet++
	}
	if hasSpecial {
		criteriaMet++
	}

	return criteriaMet >= 3
}

// 检查密码是否已泄露
func isPasswordCompromised(password string) bool {
	h := sha1.New()
	h.Write([]byte(password))
	hash := strings.ToUpper(hex.EncodeToString(h.Sum(nil)))
	prefix := hash[:5]
	suffix := hash[5:]

	// 调用HIBP API检查密码是否泄露
	resp, err := http.Get(fmt.Sprintf("https://api.pwnedpasswords.com/range/%s", prefix))
	if err != nil {
		return false
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false
	}

	// 检查后缀是否存在
	return strings.Contains(string(body), suffix)
}

func IsPasswordUsed(userID int, newPassword string) bool {
	ctx := context.Background()
	key := "pwd_history:" + strconv.Itoa(userID)

	// 获取所有历史密码
	history, err := database.RedisClient.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		log.Printf("Failed to get password history: %v", err)
		return false
	}

	// 检查新密码是否与历史密码匹配
	for _, oldHash := range history {
		if bcrypt.CompareHashAndPassword([]byte(oldHash), []byte(newPassword)) == nil {
			return true
		}
	}
	return false
}

func AddPasswordHistory(userID int, passwordHash string) {
	ctx := context.Background()
	key := "pwd_history:" + strconv.Itoa(userID)

	// 添加新密码到历史记录
	database.RedisClient.LPush(ctx, key, passwordHash)

	// 修剪列表只保留最近的几个密码
	database.RedisClient.LTrim(ctx, key, 0, MaxPasswordHistory-1)

	// 设置过期时间（30天）
	database.RedisClient.Expire(ctx, key, 30*24*time.Hour)
}
