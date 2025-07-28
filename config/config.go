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
	RabbitMQUser   string
        RabbitMQPass   string
	RabbitMQURL    string
	FromEmail      string
	SendGridAPIKey string
	IsEUAccount    bool
}

func LoadConfig() *Config {
	return &Config{
		DBUser:         getEnv("DB_USER", ""),
                DBHost:         getEnv("DB_HOST", ""),
                DBPort:         getEnv("DB_PORT", ""),
                DBPassword:     getEnv("DB_PASSWORD", ""),
                DBName:         getEnv("DB_NAME", ""),
                JWTSecret:      getEnv("JWT_SECRET", ""),
                FromEmail:      getEnv("FROM_EMAIL", ""),
                SendGridAPIKey: getEnv("SENDGRID_API_KEY", ""),
                IsEUAccount:    getEnvAsBool("SENDGRID_EU_ACCOUNT", false),
		RabbitMQUser:   getEnv("RABBITMQ_USER", ""),  
                RabbitMQPass:   getEnv("RABBITMQ_PASSWORD", ""),   
                RabbitMQURL:    fmt.Sprintf("amqp://%s:%s@%s:%s",
                                getEnv("RABBITMQ_USER", ""),
                                getEnv("RABBITMQ_PASSWORD", ""),
                                getEnv("RABBITMQ_HOST", ""),
                                getEnv("RABBITMQ_PORT", ""))}
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
