package email

import (
	"fmt"
	"net/smtp"
)

type EmailService struct {
	host     string
	port     string
	username string
	password string
}

func NewEmailService(host, port, username, password string) *EmailService {
	return &EmailService{
		host:     host,
		port:     port,
		username: username,
		password: password,
	}
}

func (s *EmailService) SendVerificationEmail(toEmail string, code string) error {
	subject := "Subject: Verify your email - AI Finance App\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; max-width: 400px; margin: auto; padding: 20px; border: 1px solid #e2e8f0; border-radius: 10px;">
			<h2 style="color: #1e293b;">Welcome to AI Finance</h2>
			<p style="color: #475569;">Your verification code is:</p>
			<h1 style="letter-spacing: 8px; color: #4F46E5; background-color: #f1f5f9; padding: 10px; text-align: center; border-radius: 5px;">%s</h1>
			<p style="color: #94a3b8; font-size: 12px;">This code expires in 15 minutes.</p>
		</div>
	`, code)

	return s.send(toEmail, subject+mime+body)
}

func (s *EmailService) SendPasswordResetEmail(toEmail string, code string) error {
	subject := "Subject: Reset your password - AI Finance App\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; max-width: 400px; margin: auto; padding: 20px; border: 1px solid #e2e8f0; border-radius: 10px;">
			<h2 style="color: #1e293b;">Password Reset Request</h2>
			<p style="color: #475569;">Your reset code is:</p>
			<h1 style="letter-spacing: 8px; color: #DC2626; background-color: #fef2f2; padding: 10px; text-align: center; border-radius: 5px;">%s</h1>
			<p style="color: #94a3b8; font-size: 12px;">This code expires in 15 minutes. If you did not request this, ignore this email.</p>
		</div>
	`, code)

	return s.send(toEmail, subject+mime+body)
}

func (s *EmailService) send(toEmail string, message string) error {
	auth := smtp.PlainAuth("", s.username, s.password, s.host)
	address := fmt.Sprintf("%s:%s", s.host, s.port)
	if err := smtp.SendMail(address, auth, s.username, []string{toEmail}, []byte(message)); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}
