package models

import (
	"time"
)

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username" binding:"omitempty,min=3"`
	Email     string    `json:"email" binding:"omitempty,email"`
	Password  string    `json:"password" binding:"required,min=6"`
	CreatedAt time.Time `json:"created_at"`
}

type LoginRequest struct {
	Identifier string `json:"identifier" binding:"required"` // 统一标识符字段
	Password   string `json:"password" binding:"required,min=6"`
}
