package main

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"net/http"
	"user-service/config"
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
		false, // auto-ack
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
			log.Printf("准备发送欢迎邮件到: %s", email)

			// 真正发送邮件
			if err := utils.SendWelcomeEmail(email); err != nil {
				log.Printf("发送失败到 %s: %v, 消息重新入队", email, err)
				err := msg.Nack(false, true)
				if err != nil {
					return
				} // 重新入队
			} else {
				log.Printf("发送成功到 %s: %v", email, err)
				err := msg.Ack(false)
				if err != nil {
					return
				}
			}
		}
	}()
}

var etcdClient *utils.EtcdClient

func main() {
	// 初始化数据库
	if err := database.InitDB(); err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	defer database.CloseDB()

	// 初始化Redis
	if err := database.InitRedis(); err != nil {
		log.Fatalf("Redis initialization failed: %v", err)
	}
	defer database.CloseRedis()

	r := gin.Default()

	// 初始化RabbitMQ
	rabbitConn, rabbitCleanup := rabbitmq.SetupRabbitMQ()
	defer rabbitCleanup()
	controllers.SetRabbitMQConnection(rabbitConn)

	// 初始化etcd
	initEtcd()
	defer deregisterEtcd()

	// 为邮件消费者创建专用通道
	ch, err := rabbitConn.Channel()
	if err != nil {
		log.Fatalf("Failed to create channel for email consumer: %v", err)
	}
	defer func(ch *amqp.Channel) {
		err := ch.Close()
		if err != nil {

		}
	}(ch)
	startEmailConsumer(ch)

	// 应用Prometheus中间件
	r.Use(middlewares.PrometheusMiddleware())

	// 应用限流中间件
	r.Use(middlewares.RateLimiter())

	// 添加Prometheus metrics端点
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// 健康检查端点
	r.GET("/health", func(c *gin.Context) {
		// 添加数据库检查
		if err := database.DB.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy"})
			return
		}
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
		protected.GET("/profile", controllers.GetUserProfile)
		protected.PUT("/password", controllers.UpdatePassword)
	}

	// 启动服务器
	log.Println("Starting server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func initEtcd() {
	cfg := config.LoadConfig()
	client, err := utils.NewEtcdClient(cfg.EtcdAddress, cfg.EtcdUsername, cfg.EtcdPassword)
	if err != nil {
		log.Printf("ETCD connection failed: %v", err)
		return
	}

	etcdClient = client

	// 修复：使用位置参数调用函数
	if err := etcdClient.RegisterService(cfg.ServiceName, cfg.ServicePort, 15); err != nil {
		log.Printf("ETCD service registration failed: %v", err)
	} else {
		log.Printf("Registered in ETCD as %s:%d", cfg.ServiceName, cfg.ServicePort)
	}
}

func deregisterEtcd() {
	if etcdClient != nil {
		if err := etcdClient.DeregisterService(); err != nil {
			log.Printf("ETCD deregister failed: %v", err)
		} else {
			log.Println("Deregistered from ETCD")
		}
		etcdClient.Close()
	}
}
