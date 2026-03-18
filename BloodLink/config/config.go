package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	MYSQL_DSN string

	JWTSECRET        string
	JWTREFRESHSECRET string
	CURR_USER        string

	SENDGRID_API_KEY string
	FROM_EMAIL       string
	FROM_NAME        string
)

func InitEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("Warning: .env file not found or could not be loaded. Relying on system environment variables.")
	}
	MYSQL_DSN = getEnv("MYSQL_DSN")

	JWTSECRET = getEnv("JWTSECRET")

	SENDGRID_API_KEY = getEnv("SENDGRID_API_KEY")
	FROM_EMAIL = getEnvWithDefault("FROM_EMAIL", "nebiyattakele23@gmail.com")
	FROM_NAME = getEnvWithDefault("FROM_NAME", "BloodLink")

	JWTREFRESHSECRET = getEnv("JWTREFRESHSECRET")

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
