package Infrastructure

import (
	"bloodlink/config"
	"fmt"
	"net/smtp"
)

func SendOTP(toEmail, otp string) error {
	from := config.FROM
	password := config.APPPASS
	smtpHost := config.SMTPSERVER
	smtpPort := config.SMTPPORT

	user := config.SMTPUSER
	auth := smtp.PlainAuth("", user, password, smtpHost)

	subject := "Subject: Verify your BloodLink account\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body := fmt.Sprintf("<html><body><strong>Your OTP is: %s</strong></body></html>", otp)
	message := []byte(subject + mime + body)

	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)
	err := smtp.SendMail(addr, auth, from, []string{toEmail}, message)
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}
	return nil
}
