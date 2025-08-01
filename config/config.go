package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	DBUser         string
	DBPassword     string
	DBHost         string
	DBPort         string
	DBName         string
	RedisHost      string
	RedisPort      string
	RedisPassword  string
	RedisDB        int
	JWTSecret      string
	RabbitMQURL    string
	FromEmail      string
	SendGridAPIKey string
	IsEUAccount    bool
	EtcdAddress    string
	EtcdUsername   string
	EtcdPassword   string
	ServiceName    string
	ServicePort    int
	RabbitMQDelay  int `json:"rabbitmq_delay"` // 延迟消息时间(毫秒)
	EtcdTTL        int `json:"etcd_ttl"`       // ETCD租约时间(秒)
	CacheTTL       int `json:"cache_ttl"`      // 缓存时间(分钟)
}

func LoadConfig() *Config {
	return &Config{
		DBUser:         getEnv("DB_USER", ""),
		DBHost:         getEnv("DB_HOST", ""),
		DBPort:         getEnv("DB_PORT", ""),
		DBPassword:     getEnv("DB_PASSWORD", ""),
		DBName:         getEnv("DB_NAME", ""),
		RedisHost:      getEnv("REDIS_HOST", ""),
		RedisPort:      getEnv("REDIS_PORT", ""),
		RedisPassword:  getEnv("REDIS_PASSWORD", ""),
		RedisDB:        getEnvAsInt("REDIS_DB", 0),
		JWTSecret:      getEnv("JWT_SECRET", ""),
		FromEmail:      getEnv("FROM_EMAIL", ""),
		SendGridAPIKey: getEnv("SENDGRID_API_KEY", ""),
		IsEUAccount:    getEnvAsBool("SENDGRID_EU_ACCOUNT", false),
		EtcdAddress:    getEnv("ETCD_ADDRESS", ""),
		EtcdUsername:   getEnv("ETCD_USERNAME", ""),
		EtcdPassword:   getEnv("ETCD_PASSWORD", ""),
		ServiceName:    getEnv("SERVICE_NAME", ""),
		ServicePort:    getEnvAsInt("SERVICE_PORT", 8080),
		RabbitMQDelay:  getEnvAsInt("RABBITMQ_DELAY", 5000),
		EtcdTTL:        getEnvAsInt("ETCD_TTL", 15),
		CacheTTL:       getEnvAsInt("CACHE_TTL", 30),
		RabbitMQURL: fmt.Sprintf("amqp://%s:%s@%s:%s", // 运行时拼接
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

func getEnvAsInt(key string, defaultValue int) int {
	value := getEnv(key, "")
	if value == "" {
		return defaultValue
	}
	num, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return num
}
