package Infrastructure

import (
	"bloodlink/config"
	"fmt"
	"log"
	"net/smtp"
	"strings"
)

func sendEmail(toEmail, subject, htmlBody string) error {
	addr := config.SMTP_HOST + ":" + config.SMTP_PORT
	auth := smtp.PlainAuth("", config.SMTP_USERNAME, config.SMTP_PASSWORD, config.SMTP_HOST)

	msg := strings.Join([]string{
		"From: " + config.FROM_NAME + " <" + config.FROM_EMAIL + ">",
		"To: " + toEmail,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=\"UTF-8\"",
		"",
		htmlBody,
	}, "\r\n")

	err := smtp.SendMail(addr, auth, config.FROM_EMAIL, []string{toEmail}, []byte(msg))
	if err != nil {
		log.Printf("[EMAIL ERROR] to=%s subject=%s error=%v", toEmail, subject, err)
		return fmt.Errorf("failed to send email: %v", err)
	}

	log.Printf("[EMAIL] Sent successfully to %s | subject=%s", toEmail, subject)
	return nil
}

func SendOTP(toEmail, otp string) error {
	subject := "Verify your BloodLink account"
	body := fmt.Sprintf("<html><body><strong>Your OTP is: %s</strong></body></html>", otp)
	return sendEmail(toEmail, subject, body)
}

func SendPasswordResetOTP(toEmail, otp string) error {
	subject := "BloodLink Password Reset"
	body := fmt.Sprintf(
		"<html><body><p>You requested a password reset for your BloodLink account.</p><p><strong>Your OTP is: %s</strong></p><p>This OTP will expire soon. If you did not request this, please ignore this email.</p></body></html>",
		otp,
	)
	return sendEmail(toEmail, subject, body)
}

func SendBloodRequestNotification(toEmail, subject, content string) error {
	body := fmt.Sprintf("<html><body><p>%s</p></body></html>", content)
	return sendEmail(toEmail, subject, body)
}
