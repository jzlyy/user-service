package config

import (
	"os"
	"strings"
)

type Config struct {
	DBUser         string
	DBPassword     string
	DBHost         string
	DBPort         string
	DBName         string
	JWTSecret      string
	RabbitMQURL    string
	FromEmail      string
	SendGridAPIKey string
	IsEUAccount    bool
}

func LoadConfig() *Config {
	return &Config{
		DBUser:         getEnv("DB_USER", "root"),
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBPort:         getEnv("DB_PORT", "3306"),
		DBName:         getEnv("DB_NAME", "ecommerce"),
		RabbitMQURL:    getEnv("RABBITMQ_URL", "amqp://admin:rabbitmq@172.168.20.30:5672/"),
		FromEmail:      getEnv("FROM_EMAIL", "wwyxhqc1@jzlyy.xyz"),
		IsEUAccount:    getEnvAsBool("SENDGRID_EU_ACCOUNT", false)}

}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return strings.TrimSpace(value) // 清理空格和换行符
	}
	return defaultValue
}

// 添加辅助函数将环境变量转为bool
func getEnvAsBool(key string, defaultValue bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	return strings.ToLower(val) == "true" || val == "1"
}
