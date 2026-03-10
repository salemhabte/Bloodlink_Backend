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
	err := godotenv.Load()
	if err != nil {
		log.Fatal("can not load .env file")
	}
	MYSQL_DSN = getEnv("MYSQL_DSN")

	JWTSECRET = getEnv("JWTSECRET")

	FROM = getEnv("FROM")
	APPPASS = getEnv("APPPASS")
	SMTPSERVER = getEnv("SMTPSERVER")
	SMTPPORT = getEnv("SMTPPORT")
	SMTPUSER = getEnv("SMTPUSER")

	JWTREFRESHSECRET = getEnv("JWTREFRESHSECRET")

}

func getEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("Environment variable %s is not set", key)
	}
	return val
}
