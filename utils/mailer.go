package utils

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"time"
	"user-service/config"
)

func SendWelcomeEmail(toEmail string) error {
	cfg := config.LoadConfig()

	// 设置带超时的拨号器
	dialer := &net.Dialer{
		Timeout: 15 * time.Second,
	}
	conn, err := dialer.Dial("tcp", fmt.Sprintf("%s:%s", cfg.SMTPHost, cfg.SMTPPort))
	if err != nil {
		log.Printf("Dial failed: %v", err)
		return err
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)

	// 创建 TLS 配置
	tlsConfig := &tls.Config{
		ServerName:         cfg.SMTPHost,
		InsecureSkipVerify: false, // 生产环境应为 false
	}

	// 创建 SMTP 客户端
	client, err := smtp.NewClient(conn, cfg.SMTPHost)
	if err != nil {
		log.Printf("SMTP client creation failed: %v", err)
		return err
	}
	defer func(client *smtp.Client) {
		err := client.Close()
		if err != nil {

		}
	}(client)

	// 启动 TLS
	if err = client.StartTLS(tlsConfig); err != nil {
		log.Printf("StartTLS failed: %v", err)
		return err
	}

	// 认证
	auth := smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPHost)
	if err = client.Auth(auth); err != nil {
		log.Printf("SMTP auth failed: %v", err)
		return err
	}

	// 设置发件人
	if err = client.Mail(cfg.FromEmail); err != nil {
		log.Printf("Mail command failed: %v", err)
		return err
	}

	// 设置收件人
	if err = client.Rcpt(toEmail); err != nil {
		log.Printf("Rcpt command failed: %v", err)
		return err
	}

	// 发送邮件内容
	w, err := client.Data()
	if err != nil {
		log.Printf("Data command failed: %v", err)
		return err
	}

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: Welcome to Our Service!\r\n\r\nHello,\r\n\r\nThank you for registering with us!", cfg.FromEmail, toEmail)
	if _, err = w.Write([]byte(msg)); err != nil {
		log.Printf("Write failed: %v", err)
		return err
	}

	if err = w.Close(); err != nil {
		log.Printf("Close writer failed: %v", err)
		return err
	}

	// 退出
	if err = client.Quit(); err != nil {
		log.Printf("Quit failed: %v", err)
		return err
	}

	log.Printf("Successfully sent welcome email to: %s", toEmail)
	return nil
}
