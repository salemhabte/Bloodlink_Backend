package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	POSTGRES_DSN string

	JWTSECRET        string
	JWTREFRESHSECRET string
	CURR_USER        string

	SMTP_HOST     string
	SMTP_PORT     string
	SMTP_USERNAME string
	SMTP_PASSWORD string
	FROM_EMAIL    string
	FROM_NAME     string
)

func InitEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("Warning: .env file not found or could not be loaded. Relying on system environment variables.")
	}

	POSTGRES_DSN = getEnv("POSTGRES_DSN")
	JWTSECRET = getEnv("JWTSECRET")
	JWTREFRESHSECRET = getEnv("JWTREFRESHSECRET")

	SMTP_HOST     = getEnvWithDefault("SMTP_HOST", "sandbox.smtp.mailtrap.io")
	SMTP_PORT     = getEnvWithDefault("SMTP_PORT", "2525")
	SMTP_USERNAME = getEnv("SMTP_USERNAME")
	SMTP_PASSWORD = getEnv("SMTP_PASSWORD")
	FROM_EMAIL    = getEnvWithDefault("FROM_EMAIL", "bloodlink@demo.com")
	FROM_NAME     = getEnvWithDefault("FROM_NAME", "BloodLink")
}

func getEnvWithDefault(key, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}

func getEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("Environment variable %s is not set", key)
	}
	return val
}
