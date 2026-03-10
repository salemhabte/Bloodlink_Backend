package Repository

import (
	"log"
)

func RunMigrations() {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS Users (
			user_id VARCHAR(36) PRIMARY KEY,
			email VARCHAR(100) UNIQUE NOT NULL,
			full_name VARCHAR(100),
			phone VARCHAR(20),
			password_hash VARCHAR(255) NOT NULL,
			role VARCHAR(50) NOT NULL,
			is_active BOOLEAN DEFAULT FALSE,
			otp VARCHAR(6),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS User_Profile (
			profile_id VARCHAR(36) PRIMARY KEY,
			user_id VARCHAR(36) NOT NULL,
			full_name VARCHAR(100),
			phone VARCHAR(20),
			city VARCHAR(100),
			area VARCHAR(100),
			profile_picture_url VARCHAR(255),
			CONSTRAINT fk_user_profile_user FOREIGN KEY (user_id) REFERENCES Users(user_id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS Donors (
			donor_id VARCHAR(36) PRIMARY KEY,
			user_id VARCHAR(36) NOT NULL,
			blood_type VARCHAR(5),
			status VARCHAR(20) DEFAULT 'Available',
			last_donation_date DATE,
			CONSTRAINT fk_donors_user FOREIGN KEY (user_id) REFERENCES Users(user_id) ON DELETE CASCADE
		);`,
	}

	for _, query := range queries {
		_, err := DB.Exec(query)
		if err != nil {
			// Ignore error 1060 (Duplicate column name) for ALTER TABLE
			if query[0:5] == "ALTER" {
				continue
			}
			log.Fatalf("Migration failed on query: %v\nError: %v", query, err)
		}
	}

	log.Println("Database tables verified/created successfully!")
}
