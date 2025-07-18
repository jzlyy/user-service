package config

import (
	"os"
	"strings"
)

type Config struct {
	DBUser      string
	DBPassword  string
	DBHost      string
	DBPort      string
	DBName      string
	JWTSecret   string
	RabbitMQURL string
	SMTPHost    string
	SMTPPort    string
	SMTPUser    string
	SMTPPass    string
	FromEmail   string
}

func LoadConfig() *Config {
	return &Config{
		DBUser:      getEnv("DB_USER", "root"),
		DBPassword:  getEnv("DB_PASSWORD", "qzhufuchengz"),
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnv("DB_PORT", "3306"),
		DBName:      getEnv("DB_NAME", "ecommerce"),
		JWTSecret:   getEnv("JWT_SECRET", "G9mCQ19ogTkuWQY9jH2wGZASuGi/JrhstQaZy4k/01o="),
		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://admin:rabbitmq@IP:5672/"),
		SMTPHost:    getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:    getEnv("SMTP_PORT", "587"),
		SMTPUser:    getEnv("SMTP_USER", ""),
		SMTPPass:    getEnv("SMTP_PASS", ""),
		FromEmail:   getEnv("FROM_EMAIL", "")}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return strings.TrimSpace(value) // 清理空格和换行符
	}
	return defaultValue
}
