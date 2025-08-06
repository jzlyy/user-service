package controllers

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"time"
	"user-service/config"
	"user-service/database"
	"user-service/models"
	"user-service/rabbitmq"
	"user-service/utils"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	"golang.org/x/crypto/bcrypt"
)

// RefreshTokenRequest 刷新令牌请求结构体
type RefreshTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

// Register godoc
// @Summary 注册新用户
// @Description 注册一个新用户
// @Tags auth
// @Accept  json
// @Produce  json
// @Param   user body models.User true "用户信息"
// @Success 201 {object} map[string]interface{} "创建成功"
// @Failure 400 {object} map[string]string "请求错误"
// @Failure 500 {object} map[string]string "服务器错误"
// @Router /register [post]
func Register(c *gin.Context) {
	// 添加配置加载
	cfg := config.LoadConfig()

	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 密码强度检查
	if !utils.ValidatePasswordStrength(user.Password) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "密码必须至少8位，且包含大写字母、小写字母、数字和特殊字符中的三种",
		})
		return
	}

	// 确保邮箱地址被正确接收
	if user.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not hash password"})
		return
	}

	result, err := database.DB.Exec(
		"INSERT INTO users (username, email, password, created_at) VALUES (?, ?, ?, ?)",
		user.Username,
		user.Email,
		string(hashedPassword),
		time.Now(),
	)
	if err != nil {
		log.Printf("[AUDIT] Failed to register user: %s, error: %v", user.Email, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create user"})
		return
	}

	userID, _ := result.LastInsertId()
	log.Printf("[AUDIT] User registered: ID=%d, Email=%s", userID, user.Email)
	c.JSON(http.StatusCreated, gin.H{"id": userID, "message": "User created"})

	// 注册成功后发送延迟消息
	ch, _ := rabbitmq.GetChannel()
	if ch == nil {
		log.Printf("Failed to get RabbitMQ channel")
		return
	}
	defer rabbitmq.ReleaseChannel(ch)

	// 加载配置获取RabbitMQ延迟
	cfg = config.LoadConfig()

	msg := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		ContentType:  "text/plain",
		Body:         []byte(user.Email),
		Headers: amqp.Table{
			"x-delay": cfg.RabbitMQDelay, // 使用配置值
		},
	}

	err = ch.Publish(
		rabbitmq.DelayedExchange,
		"", // routing key
		false,
		false,
		msg,
	)

	if err != nil {
		log.Printf("Failed to send welcome email: %v", err)
	}
}

// Login godoc
// @Summary 用户登录
// @Description 用户登录并获取Token
// @Tags auth
// @Accept  json
// @Produce  json
// @Param   credentials body models.LoginRequest true "登录凭证"
// @Success 200 {object} map[string]string "登录成功"
// @Failure 400 {object} map[string]string "请求错误"
// @Failure 401 {object} map[string]string "认证失败"
// @Router /login [post]
func Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查是邮箱还是用户名
	var user models.User
	var queryField string
	var queryValue string

	if utils.IsEmail(req.Identifier) {
		queryField = "email"
		queryValue = req.Identifier
	} else {
		queryField = "username"
		queryValue = req.Identifier
	}

	err := database.DB.QueryRow(
		"SELECT id, password FROM users WHERE "+queryField+" = ?",
		queryValue,
	).Scan(&user.ID, &user.Password)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		log.Printf("[AUDIT] Failed login attempt: %s", req.Identifier)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	log.Printf("[AUDIT] User logged in: ID=%d", user.ID)

	//生成Token
	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// RefreshToken godoc
// @Summary 刷新Token
// @Description 使用旧Token获取新Token
// @Tags auth
// @Accept  json
// @Produce  json
// @Param   request body RefreshTokenRequest true "刷新令牌请求"
// @Success 200 {object} map[string]string "刷新成功"
// @Failure 400 {object} map[string]string "请求错误"
// @Failure 401 {object} map[string]string "认证失败"
// @Router /refresh-token [post]
// RefreshToken 添加新的路由处理函数
func RefreshToken(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	newToken, err := utils.RefreshToken(req.Token)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": newToken})
}

func ProtectedEndpoint(c *gin.Context) {
	userID, _ := c.Get("userID")
	c.JSON(http.StatusOK, gin.H{"message": "Access granted", "user_id": userID})
}
