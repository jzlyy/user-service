package main

import (
	"log"
	"user-service/rabbitmq"

	amqp "github.com/rabbitmq/amqp091-go"
)

func startEmailConsumer(ch *amqp.Channel) {
	msgs, err := ch.Consume(
		rabbitmq.WelcomeEmailQueue,
		"",    // consumers
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to register consumers: %v", err)
	}

	go func() {
		for msg := range msgs {
			email := string(msg.Body)
			log.Printf("Sending welcome email to: %s", email)
			// 实际邮件发送逻辑
			// sendWelcomeEmail(email)
		}
	}()
}
