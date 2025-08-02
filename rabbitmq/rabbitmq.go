package rabbitmq

import (
	"log"
	"sync"
	"user-service/config"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	WelcomeEmailQueue = "welcome_email_queue"
	DelayedExchange   = "delayed_exchange"
)

var (
	connection  *amqp.Connection
	channelPool *sync.Pool
)

func SetupRabbitMQ() (*amqp.Connection, func()) {
	cfg := config.LoadConfig()

	var err error
	connection, err = amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	// 初始化通道池
	channelPool = &sync.Pool{
		New: func() interface{} {
			ch, err := connection.Channel()
			if err != nil {
				log.Printf("Failed to create channel: %v", err)
				return nil
			}
			return ch
		},
	}

	// 获取临时通道进行初始化
	initCh, err := connection.Channel()
	if err != nil {
		log.Fatalf("Failed to open channel: %v", err)
	}
	defer func(initCh *amqp.Channel) {
		err := initCh.Close()
		if err != nil {
			log.Printf("Failed to close channel: %v", err)
		}
	}(initCh)

	// 声明延迟交换机
	err = initCh.ExchangeDeclare(
		DelayedExchange,
		"x-delayed-message",
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		amqp.Table{"x-delayed-type": "direct"},
	)
	if err != nil {
		log.Fatalf("Failed to declare delayed exchange: %v", err)
	}

	// 声明队列
	_, err = initCh.QueueDeclare(
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
	err = initCh.QueueBind(
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
		if connection != nil {
			err := connection.Close()
			if err != nil {
				return
			}
		}
	}

	return connection, cleanup
}

// GetChannel 获取通道
func GetChannel() *amqp.Channel {
	return channelPool.Get().(*amqp.Channel)
}

// ReleaseChannel 释放通道
func ReleaseChannel(ch *amqp.Channel) {
	if ch != nil {
		channelPool.Put(ch)
	}
}
