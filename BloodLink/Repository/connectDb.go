package Repository

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"bloodlink/config"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func ConnectDB() {
	dsn := config.POSTGRES_DSN
	var err error
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Error opening database: ", err)
	}

	// Test the connection
	err = DB.Ping()
	if err != nil {
		log.Fatal("Error connecting to database: ", err)
	}
	DB.SetMaxOpenConns(10)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(5 * time.Minute)

	fmt.Println("Connected to PostgreSQL database successfully!")
}
