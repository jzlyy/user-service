package controllers

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"user-service/database"
	"user-service/services"
	"user-service/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// GetUserProfile godoc
// @Summary 获取用户资料
// @Description 获取当前登录用户的资料
// @Tags user
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {object} models.User "成功获取"
// @Failure 401 {object} map[string]string "未授权"
// @Failure 404 {object} map[string]string "用户不存在"
// @Router /profile [get]
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

// UpdatePassword godoc
// @Summary 更新密码
// @Description 更新当前用户的密码
// @Tags user
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param   request body UpdatePasswordRequest true "密码更新请求"
// @Success 200 {object} map[string]string "更新成功"
// @Failure 400 {object} map[string]string "请求错误"
// @Failure 401 {object} map[string]string "未授权"
// @Router /password [put]
func UpdatePassword(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	// 验证新密码强度
	if !utils.ValidatePasswordStrength(req.NewPassword) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "密码必须至少8位，且包含大写字母、小写字母、数字和特殊字符中的三种",
		})
		return
	}

	var currentPassword string
	err := database.DB.QueryRow("SELECT password FROM users WHERE id = ?", userID).Scan(&currentPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			log.Printf("Database error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(currentPassword), []byte(req.OldPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid old password"})
		return
	}

	// 检查新密码是否与旧密码相同
	if bcrypt.CompareHashAndPassword([]byte(currentPassword), []byte(req.NewPassword)) == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "New password cannot be same as old password"})
		return
	}

	// 检查新密码是否使用过
	if utils.IsPasswordUsed(userID.(int), req.NewPassword) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Cannot reuse recent passwords"})
		return
	}

	// 生成新密码哈希
	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Password hashing error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not hash password"})
		return
	}

	// 更新数据库
	_, err = database.DB.Exec("UPDATE users SET password = ? WHERE id = ?", string(newHash), userID)
	if err != nil {
		log.Printf("Password update error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update password"})
		return
	}

	// 添加密码到历史记录
	utils.AddPasswordHistory(userID.(int), string(newHash))

	// 更新成功后删除缓存
	go func() {
		if err := utils.InvalidateUserCache(userID.(int)); err != nil {
			log.Printf("Cache invalidation error: %v", err)
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}
