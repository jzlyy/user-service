package rabbitmq

import (
	"log"
	"user-service/config"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	WelcomeEmailQueue = "welcome_email_queue"
	DelayedExchange   = "delayed_exchange"
)

func SetupRabbitMQ() (*amqp.Channel, func()) {
	cfg := config.LoadConfig()

	conn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open channel: %v", err)
	}

	// 声明延迟交换机
	err = ch.ExchangeDeclare(
		DelayedExchange,
		"x-delayed-message", // 特殊交换机类型
		true,                // durable
		false,               // auto-deleted
		false,               // internal
		false,               // no-wait
		amqp.Table{"x-delayed-type": "direct"},
	)
	if err != nil {
		log.Fatalf("Failed to declare delayed exchange: %v", err)
	}

	// 声明队列
	_, err = ch.QueueDeclare(
		WelcomeEmailQueue,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}

	// 绑定队列到交换机
	err = ch.QueueBind(
		WelcomeEmailQueue,
		"", // routing key
		DelayedExchange,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to bind queue: %v", err)
	}

	cleanup := func() {
		err := ch.Close()
		if err != nil {
			return
		}
		err = conn.Close()
		if err != nil {
			return
		}
	}

	return ch, cleanup
}
