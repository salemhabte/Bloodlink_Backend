package config

import (
	"log"
	"os"
	

	"github.com/joho/godotenv"
)

var (
	MONGO_CONNECTION_STRING string
	// GEMINI_API_KEY          string

	USER_DB              string
	USER_COLLECTION_NAME string

	
	USER_OTP_COLLECTION_NAME           string
	USER_REFRESH_TOKEN_COLLECTION_NAME string

	JWTSECRET        string
	JWTREFRESHSECRET string
	CURR_USER        string

    MysqlDSN string

	FROM       string
	APPPASS    string
	SMTPSERVER string
	SMTPPORT   string
	SMTPUSER   string

	CLIENT_ID           string
	CLIENT_SECRET       string
	CLIENT_CALLBACK_URL string
)

func InitEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("can not load .env file")
	}
	MONGO_CONNECTION_STRING = getEnv("MONGO_CONNECTION_STRING")
	// GEMINI_API_KEY = getEnv("GEMINI_API_KEY")
	USER_DB = getEnv("USER_DB")
	USER_COLLECTION_NAME = getEnv("USER_COLLECTION_NAME")
	
	JWTSECRET = getEnv("JWTSECRET")
	CURR_USER = getEnv("CURR_USER")
	
	FROM = getEnv("FROM")
	APPPASS = getEnv("APPPASS")
	SMTPSERVER = getEnv("SMTPSERVER")
	SMTPPORT = getEnv("SMTPPORT")
	SMTPUSER = getEnv("SMTPUSER")
	CLIENT_ID = getEnv("CLIENT_ID")
	CLIENT_SECRET = getEnv("CLIENT_SECRET")
	CLIENT_CALLBACK_URL = getEnv("CLIENT_CALLBACK_URL")
	USER_OTP_COLLECTION_NAME = getEnv("USER_OTP_COLLECTION_NAME")
	USER_REFRESH_TOKEN_COLLECTION_NAME = getEnv("USER_REFRESH_TOKEN_COLLECTION_NAME")
	JWTREFRESHSECRET = getEnv("JWTREFRESHSECRET")
	
	
}

func getEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("Environment variable %s is not set", key)
	}
	return val
}

type Config struct {
    MysqlDSN string
}

func LoadConfig() Config {
    InitEnv()
    return Config{
        MysqlDSN: getEnv("MYSQL_DSN"),
    }
}