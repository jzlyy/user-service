package controllers

import (
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"user-service/database"
	"user-service/services"
	"user-service/utils"
)

func GetUserProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// 尝试从缓存获取
	var cachedUser struct {
		ID        int    `json:"id"`
		Username  string `json:"username"`
		Email     string `json:"email"`
		CreatedAt string `json:"created_at"`
	}

	if found, err := utils.GetUserCache(userID.(int), &cachedUser); err == nil && found {
		c.JSON(http.StatusOK, cachedUser)
		return
	}

	// 使用服务层查询函数
	user, err := services.GetUserByID(userID.(int))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	// 更新缓存（异步）
	go func() {
		if err := utils.SetUserCache(user.ID, user); err != nil {
			log.Printf("Failed to cache user: %v", err)
		}
	}()

	c.JSON(http.StatusOK, user)
}

func UpdatePassword(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证旧密码
	var currentPassword string
	err := database.DB.QueryRow("SELECT password FROM users WHERE id = ?", userID).Scan(&currentPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(currentPassword), []byte(req.OldPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid old password"})
		return
	}

	// 检查新密码是否使用过
	if utils.IsPasswordUsed(userID.(int), req.NewPassword) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Cannot reuse recent passwords"})
		return
	}

	// 更新密码
	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not hash password"})
		return
	}

	_, err = database.DB.Exec("UPDATE users SET password = ? WHERE id = ?", string(newHash), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update password"})
		return
	}

	// 添加密码到历史记录
	utils.AddPasswordHistory(userID.(int), string(newHash))

	// 更新成功后删除缓存
	go func() {
		if err := utils.InvalidateUserCache(userID.(int)); err != nil {
			log.Printf("Failed to invalidate user cache: %v", err)
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}
