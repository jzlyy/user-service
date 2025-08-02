package database

import (
	"database/sql"
	"errors"
	"fmt"
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
