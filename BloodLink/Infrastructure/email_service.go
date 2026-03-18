package Infrastructure

import (
	"bloodlink/Domain"
	"bloodlink/config"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func sendGridEmail(toEmail, subject, htmlBody string) error {
	reqBody := Domain.SendGridRequest{
		Personalizations: []Domain.Personalization{
			{
				To: []Domain.EmailAddress{
					{Email: toEmail},
				},
			},
		},
		From: Domain.EmailAddress{
			Email: config.FROM_EMAIL,
			Name:  config.FROM_NAME,
		},
		Subject: subject,
		Content: []Domain.Content{
			{
				Type:  "text/html",
				Value: htmlBody,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal sendgrid request: %v", err)
	}

	req, err := http.NewRequest("POST", "https://api.sendgrid.com/v3/mail/send", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+config.SENDGRID_API_KEY)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to sendgrid: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("sendgrid api error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func SendOTP(toEmail, otp string) error {
	subject := "Verify your BloodLink account"
	body := fmt.Sprintf("<html><body><strong>Your OTP is: %s</strong></body></html>", otp)
	return sendGridEmail(toEmail, subject, body)
}

func SendPasswordResetOTP(toEmail, otp string) error {
	subject := "BloodLink Password Reset"
	body := fmt.Sprintf("<html><body><p>You requested a password reset for your BloodLink account.</p><p><strong>Your OTP is: %s</strong></p><p>This OTP will expire soon. If you did not request this, please ignore this email.</p></body></html>", otp)
	return sendGridEmail(toEmail, subject, body)
}
