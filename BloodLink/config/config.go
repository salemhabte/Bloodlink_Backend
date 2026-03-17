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

	FROM       string
	APPPASS    string
	SMTPSERVER string
	SMTPPORT   string
	SMTPUSER   string
)

func InitEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("Warning: .env file not found or could not be loaded. Relying on system environment variables.")
	}
	MYSQL_DSN = getEnv("MYSQL_DSN")

	JWTSECRET = getEnv("JWTSECRET")

	FROM = getEnv("FROM")
	APPPASS = getEnv("APPPASS")
	SMTPSERVER = getEnv("SMTPSERVER")
	SMTPPORT = getEnvWithDefault("SMTPPORT", "2525")
	SMTPUSER = getEnv("SMTPUSER")

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
