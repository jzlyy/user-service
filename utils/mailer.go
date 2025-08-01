package utils

import (
	"errors"
	"fmt"
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

	// 优化后的HTML内容
	htmlContent := `
<!DOCTYPE html>
<html>
<head>
    <meta http-equiv="Content-Type" content="text/html; charset=utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Welcome to Our Service!</title>
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            line-height: 1.6;
            color: #333;
            background-color: #f9f9f9;
            margin: 0;
            padding: 0;
        }
        .container {
            max-width: 600px;
            margin: 20px auto;
            background: #ffffff;
            border-radius: 8px;
            overflow: hidden;
            box-shadow: 0 0 20px rgba(0, 0, 0, 0.1);
        }
        .header {
            background: linear-gradient(135deg, #6a11cb 0%, #2575fc 100%);
            padding: 30px 20px;
            text-align: center;
        }
        .header h1 {
            color: white;
            margin: 0;
            font-size: 28px;
            font-weight: 600;
        }
        .content {
            padding: 30px;
        }
        .welcome-text {
            font-size: 18px;
            margin-bottom: 25px;
            color: #444;
        }
        .highlight {
            background-color: #f0f7ff;
            border-left: 4px solid #2575fc;
            padding: 15px;
            margin: 20px 0;
            border-radius: 0 4px 4px 0;
        }
        .cta-button {
            display: inline-block;
            background: #2575fc;
            color: white !important;
            text-decoration: none;
            padding: 12px 25px;
            border-radius: 4px;
            font-weight: 600;
            margin: 20px 0;
            text-align: center;
        }
        .features {
            display: flex;
            flex-wrap: wrap;
            margin: 25px 0;
        }
        .feature {
            flex: 1;
            min-width: 150px;
            text-align: center;
            padding: 15px;
            margin: 10px;
            background: #f8f9fa;
            border-radius: 8px;
        }
        .feature-icon {
            font-size: 24px;
            margin-bottom: 10px;
            color: #2575fc;
        }
        .footer {
            text-align: center;
            padding: 20px;
            background: #f1f5f9;
            color: #666;
            font-size: 14px;
        }
        .social-links {
            margin: 15px 0;
        }
        .social-links a {
            display: inline-block;
            margin: 0 10px;
            color: #2575fc;
            text-decoration: none;
        }
        @media (max-width: 480px) {
            .container {
                margin: 10px;
            }
            .content {
                padding: 20px;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Welcome to Our Service!</h1>
        </div>
        
        <div class="content">
            <p class="welcome-text">Hello,</p>
            <p class="welcome-text">Thank you for registering with us! We're excited to have you on board.</p>
            
            <div class="highlight">
                <p>Your account is now active and ready to use. We've prepared everything you need to get started.</p>
            </div>
            
            <div style="text-align: center;">
                <a href="#" class="cta-button">Get Started</a>
            </div>
            
            <h3 style="color: #2575fc; margin-top: 30px;">What You Can Do Now:</h3>
            <div class="features">
                <div class="feature">
                    <div class="feature-icon">🔐</div>
                    <h4>Secure Login</h4>
                    <p>Access your account anytime with enhanced security</p>
                </div>
                <div class="feature">
                    <div class="feature-icon">⚙️</div>
                    <h4>Custom Settings</h4>
                    <p>Personalize your experience</p>
                </div>
                <div class="feature">
                    <div class="feature-icon">📱</div>
                    <h4>Mobile Access</h4>
                    <p>Use our service on any device</p>
                </div>
            </div>
            
            <p>If you have any questions, feel free to reply to this email. Our support team is here to help!</p>
        </div>
        
        <div class="footer">
            <div class="social-links">
                <a href="#">Facebook</a> • 
                <a href="#">Twitter</a> • 
                <a href="#">LinkedIn</a>
            </div>
            <p>© 2025 Our Service. All rights reserved.</p>
            <p>123 Business Street, Tech City, TC 10001</p>
            <p><a href="#" style="color: #2575fc; text-decoration: none;">Unsubscribe</a> | 
            <a href="#" style="color: #2575fc; text-decoration: none;">Privacy Policy</a></p>
        </div>
    </div>
</body>
</html>
	`

	// 纯文本内容
	plainTextContent := `Hello,

Thank you for registering with us! We're excited to have you on board.

Your account is now active and ready to use. We've prepared everything you need to get started.

What You Can Do Now:
- Secure Login: Access your account anytime with enhanced security
- Custom Settings: Personalize your experience
- Mobile Access: Use our service on any device

If you have any questions, feel free to reply to this email. Our support team is here to help!

Best regards,
Service Team`

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

	errorMsg := fmt.Sprintf("SendGrid error [%d]: %s", response.StatusCode, response.Body)
	log.Printf(errorMsg)
	return errors.New(errorMsg)
}
