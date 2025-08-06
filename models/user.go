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

type ErrorResponse struct {
	Error string `json:"error" example:"错误信息"`
}

type MessageResponse struct {
	Message string `json:"message" example:"成功信息"`
}

type TokenResponse struct {
	Token string `json:"token" example:"JWT令牌"`
}
