package utils

import (
	"errors"
	"log"
	"user-service/config"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func SendWelcomeEmail(toEmail string) error {
	cfg := config.LoadConfig()

	// 创建邮件对象
	from := mail.NewEmail("Service Team", cfg.FromEmail)
	to := mail.NewEmail("New User", toEmail)
	subject := "Welcome to Our Service!"
	plainTextContent := "Hello,\n\nThank you for registering with us!"
	htmlContent := "<strong>Hello,</strong><br><br>Thank you for registering with us!"

	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)

	// 创建SendGrid客户端
	client := sendgrid.NewSendClient(cfg.SendGridAPIKey)

	// 根据配置决定是否设置欧盟数据驻留
	if cfg.IsEUAccount {
		client.Request, _ = sendgrid.SetDataResidency(client.Request, "eu")
	}

	// 发送邮件
	response, err := client.Send(message)
	if err != nil {
		log.Printf("Failed to send welcome email: %v", err)
		return err
	}

	if response.StatusCode >= 200 && response.StatusCode < 300 {
		log.Printf("Successfully sent welcome email to: %s", toEmail)
		return nil
	}

	log.Printf("SendGrid returned error: %d %s ", response.StatusCode, response.Body)
	return errors.New("sendgrid error: " + response.Body)
}
