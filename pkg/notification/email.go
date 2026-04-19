package notification

import (
	"fmt"
	"net/smtp"
	"strings"
	"time"
)

type EmailConfig struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
}

type EmailService struct {
	config EmailConfig
}

func NewEmailService(config EmailConfig) *EmailService {
	return &EmailService{config: config}
}

func (e *EmailService) SendOTP(to, otp string) error {
	subject := "Your OTP Code"
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; background-color: #f4f4f4; margin: 0; padding: 20px; }
        .container { max-width: 600px; margin: 0 auto; background-color: #ffffff; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .header { text-align: center; color: #333; }
        .otp-code { font-size: 32px; font-weight: bold; color: #4CAF50; text-align: center; padding: 20px; background-color: #f9f9f9; border-radius: 4px; margin: 20px 0; letter-spacing: 8px; }
        .message { color: #666; line-height: 1.6; }
        .footer { margin-top: 20px; padding-top: 20px; border-top: 1px solid #eee; color: #999; font-size: 12px; text-align: center; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2>OTP Verification</h2>
        </div>
        <div class="message">
            <p>Hello,</p>
            <p>Your One-Time Password (OTP) for verification is:</p>
        </div>
        <div class="otp-code">%s</div>
        <div class="message">
            <p>This code will expire in 5 minutes.</p>
            <p>If you didn't request this code, please ignore this email.</p>
        </div>
        <div class="footer">
            <p>This is an automated message, please do not reply.</p>
        </div>
    </div>
</body>
</html>
`, otp)

	return e.sendEmail(to, subject, body)
}

func (e *EmailService) SendPasswordReset(to, resetLink string) error {
	subject := "Password Reset Request"
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; background-color: #f4f4f4; margin: 0; padding: 20px; }
        .container { max-width: 600px; margin: 0 auto; background-color: #ffffff; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .header { text-align: center; color: #333; }
        .message { color: #666; line-height: 1.6; }
        .button { display: inline-block; padding: 12px 24px; margin: 20px 0; background-color: #4CAF50; color: #ffffff; text-decoration: none; border-radius: 4px; }
        .footer { margin-top: 20px; padding-top: 20px; border-top: 1px solid #eee; color: #999; font-size: 12px; text-align: center; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2>Password Reset</h2>
        </div>
        <div class="message">
            <p>Hello,</p>
            <p>We received a request to reset your password. Click the button below to reset it:</p>
            <p style="text-align: center;">
                <a href="%s" class="button">Reset Password</a>
            </p>
            <p>Or copy and paste this link into your browser:</p>
            <p style="word-break: break-all; color: #4CAF50;">%s</p>
            <p>This link will expire in 1 hour.</p>
            <p>If you didn't request a password reset, please ignore this email.</p>
        </div>
        <div class="footer">
            <p>This is an automated message, please do not reply.</p>
        </div>
    </div>
</body>
</html>
`, resetLink, resetLink)

	return e.sendEmail(to, subject, body)
}

func (e *EmailService) SendWelcome(to, username string) error {
	subject := "Welcome to GoConnect!"
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; background-color: #f4f4f4; margin: 0; padding: 20px; }
        .container { max-width: 600px; margin: 0 auto; background-color: #ffffff; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .header { text-align: center; color: #333; }
        .message { color: #666; line-height: 1.6; }
        .footer { margin-top: 20px; padding-top: 20px; border-top: 1px solid #eee; color: #999; font-size: 12px; text-align: center; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2>Welcome to GoConnect!</h2>
        </div>
        <div class="message">
            <p>Hello %s,</p>
            <p>Thank you for joining GoConnect! We're excited to have you on board.</p>
            <p>You can now access all the features of our platform.</p>
            <p>If you have any questions, feel free to reach out to our support team.</p>
        </div>
        <div class="footer">
            <p>This is an automated message, please do not reply.</p>
        </div>
    </div>
</body>
</html>
`, username)

	return e.sendEmail(to, subject, body)
}

func (e *EmailService) sendEmail(to, subject, body string) error {
	auth := smtp.PlainAuth("", e.config.SMTPUsername, e.config.SMTPPassword, e.config.SMTPHost)

	from := fmt.Sprintf("%s <%s>", e.config.FromName, e.config.FromEmail)
	
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	addr := fmt.Sprintf("%s:%s", e.config.SMTPHost, e.config.SMTPPort)
	
	return smtp.SendMail(addr, auth, e.config.FromEmail, []string{to}, []byte(message))
}

func (e *EmailService) SendBulk(recipients []string, subject, body string) error {
	var errors []string
	
	for _, recipient := range recipients {
		if err := e.sendEmail(recipient, subject, body); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", recipient, err))
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("failed to send to some recipients: %s", strings.Join(errors, "; "))
	}
	
	return nil
}

func (e *EmailService) SendBlockNotification(to string, blockDuration time.Duration) error {
	subject := "Account Temporarily Blocked - Too Many Failed Attempts"
	minutes := int(blockDuration.Minutes())
	
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; background-color: #f4f4f4; margin: 0; padding: 20px; }
        .container { max-width: 600px; margin: 0 auto; background-color: #ffffff; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .header { text-align: center; color: #d32f2f; }
        .warning-icon { font-size: 48px; text-align: center; color: #ff9800; margin: 20px 0; }
        .message { color: #666; line-height: 1.6; }
        .block-time { font-size: 24px; font-weight: bold; color: #d32f2f; text-align: center; padding: 20px; background-color: #fff3e0; border-radius: 4px; margin: 20px 0; }
        .footer { margin-top: 20px; padding-top: 20px; border-top: 1px solid #eee; color: #999; font-size: 12px; text-align: center; }
        ul { padding-left: 20px; }
        li { margin: 10px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="warning-icon">⚠️</div>
        <div class="header">
            <h2>Account Temporarily Blocked</h2>
        </div>
        <div class="message">
            <p>Hello,</p>
            <p>Your account has been temporarily blocked due to too many failed verification attempts.</p>
            <p><strong>This is a security measure to protect your account.</strong></p>
        </div>
        <div class="block-time">Blocked for %d minutes</div>
        <div class="message">
            <p>You can try again after the block period expires.</p>
            <p>If you did not attempt to verify your email, please ignore this message or contact support if you have concerns.</p>
            <p><strong>Security Tips:</strong></p>
            <ul>
                <li>Never share your verification codes with anyone</li>
                <li>Always verify the sender email is from GoConnect</li>
                <li>Contact support if you suspect unauthorized access</li>
            </ul>
        </div>
        <div class="footer">
            <p>This is an automated security notification.</p>
        </div>
    </div>
</body>
</html>
`, minutes)

	return e.sendEmail(to, subject, body)
}
