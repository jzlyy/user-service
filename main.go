package main

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"net/http"
	"user-service/controllers"
	"user-service/database"
	"user-service/middlewares"
	"user-service/rabbitmq"
	"user-service/utils"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// 添加邮件消费者实现
func startEmailConsumer(ch *amqp.Channel) {
	msgs, err := ch.Consume(
		rabbitmq.WelcomeEmailQueue,
		"",    // consumer
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to register consumer: %v", err)
	}

	go func() {
		for msg := range msgs {
			email := string(msg.Body)
			log.Printf("Sending welcome email to: %s", email)

			// 真正发送邮件
			if err := utils.SendWelcomeEmail(email); err != nil {
				log.Printf("Failed to send welcome email: %v", err)
			} else {
				log.Printf("Successfully sent welcome email to %s", email)
			}
		}
	}()
}

func main() {
	// 初始化数据库
	if err := database.InitDB(); err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	defer database.CloseDB()

	r := gin.Default()

	// 初始化RabbitMQ
	rabbitCh, rabbitCleanup := rabbitmq.SetupRabbitMQ()
	defer rabbitCleanup()
	controllers.SetRabbitMQChannel(rabbitCh)

	// 启动邮件消费者
	startEmailConsumer(rabbitCh)

	// 应用Prometheus中间件
	r.Use(middlewares.PrometheusMiddleware())

	// 添加Prometheus metrics端点
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// 健康检查端点
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// 公共路由
	public := r.Group("/api")
	{
		public.POST("/register", controllers.Register)
		public.POST("/login", controllers.Login)
	}

	// 受保护路由
	protected := r.Group("/api")
	protected.Use(middlewares.AuthMiddleware())
	{
		protected.GET("/protected", controllers.ProtectedEndpoint)
	}

	// 启动服务器
	log.Println("Starting server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
