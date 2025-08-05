package main

import (
	"context"
	"errors"
	"github.com/opentracing/opentracing-go"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"user-service/config"
	"user-service/controllers"
	"user-service/database"
	"user-service/middlewares"
	"user-service/rabbitmq"
	"user-service/utils"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var nacosClient *utils.NacosClient

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

	// 注册错误处理中间件
	r.Use(middlewares.ErrorHandler())

	// 初始化RabbitMQ
	rabbitConn, rabbitCleanup := rabbitmq.SetupRabbitMQ()
	defer rabbitCleanup()

	// 初始化Nacos
	initNacos()
	defer deregisterNacos()

	// 初始化追踪
	tracer, closer, err := utils.InitTracing("user-service")
	if err != nil {
		log.Fatalf("Failed to init tracing: %v", err)
	}
	defer closer()
	opentracing.SetGlobalTracer(tracer)

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

	//应用链路追踪中间件
	r.Use(middlewares.TracingMiddleware())

	// 健康检查端点
	r.GET("/health", func(c *gin.Context) {
		status := http.StatusOK
		components := make(map[string]string)

		// 数据库检查
		if err := database.DB.Ping(); err != nil {
			status = http.StatusServiceUnavailable
			components["database"] = "down"
		} else {
			components["database"] = "ok"
		}

		// Redis检查
		if _, err := database.RedisClient.Ping(context.Background()).Result(); err != nil {
			status = http.StatusServiceUnavailable
			components["redis"] = "down"
		} else {
			components["redis"] = "ok"
		}

		// RabbitMQ检查
		ch, err := rabbitmq.GetChannel()
		if err != nil {
			status = http.StatusServiceUnavailable
			components["rabbitmq"] = "down"
		} else {
			components["rabbitmq"] = "ok"
			rabbitmq.ReleaseChannel(ch)
		}

		c.JSON(status, gin.H{
			"status":     http.StatusText(status),
			"components": components,
		})
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
		protected.POST("/refresh-token", controllers.RefreshToken)
	}

	// 建立连接
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// 创建优雅关闭信号通道
	shutdown := make(chan struct{})

	// 启动服务器
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
		close(shutdown)
	}()

	// 等待关闭信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		log.Println("Shutdown initiated...")
	case <-shutdown:
		log.Println("Server closed unexpectedly")
	}

	// 分阶段关闭
	phase1 := make(chan struct{})
	go func() {
		defer close(phase1)

		// 第一阶段：关闭外部依赖
		deregisterNacos()
		rabbitCleanup()

		// 停止接受新请求
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		err := srv.Shutdown(ctx)
		if err != nil {
			return
		}
	}()

	// 第二阶段：关闭数据库连接
	select {
	case <-phase1:
	case <-time.After(20 * time.Second):
	}

	database.CloseDB()
	database.CloseRedis()

	log.Println("Server exiting")
}

func initNacos() {
	cfg := config.LoadConfig()
	client, err := utils.NewNacosClient(cfg)
	if err != nil {
		log.Printf("Nacos initialization failed: %v", err)
		return
	}
	nacosClient = client // 赋值给全局变量

	if err := nacosClient.RegisterService(); err != nil {
		log.Printf("Nacos service registration failed: %v", err)
	} else {
		log.Printf("Registered in Nacos as %s:%d", cfg.ServiceName, cfg.ServicePort)
	}
}

func deregisterNacos() {
	if nacosClient != nil {
		if err := nacosClient.DeregisterService(); err != nil {
			log.Printf("Nacos deregister failed: %v", err)
		} else {
			log.Println("Deregistered from Nacos")
		}
	}
}
