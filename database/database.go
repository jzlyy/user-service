package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"
	"user-service/config"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB
var (
	RedisError = errors.New("redis error")
)

func InitDB() error {
	cfg := config.LoadConfig()
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}

	if err = db.Ping(); err != nil {
		return err
	}
	DB = db

	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(10 * time.Minute)

	// 健康检查协程
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			if err := db.Ping(); err != nil {
				log.Printf("Database connection unhealthy: %v", err)

				// 尝试恢复连接
				if err := db.Close(); err == nil {
					if newDB, err := sql.Open("mysql", dsn); err == nil {
						db = newDB
					}
				}
			}
		}
	}()

	return nil
}

func CloseDB() {
	if DB != nil {
		err := DB.Close()
		if err != nil {
			return
		}
	}
}
