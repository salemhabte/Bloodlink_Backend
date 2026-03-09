package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"bloodlink/Repository" // Make sure this matches your folder name exactly
)

func main() {
	// Step 1: Connect to database
	Repository.ConnectDB()
	db := Repository.DB
	defer db.Close()

	// Step 2: Path to migration folder
	migrationFolder := "migrations"

	// Step 3: Get all .sql files in order
	files, err := filepath.Glob(filepath.Join(migrationFolder, "*.sql"))
	if err != nil {
		log.Fatal("Error reading migration folder:", err)
	}
	if len(files) == 0 {
		log.Fatal("No migration files found in folder:", migrationFolder)
	}

	// Step 4: Run each migration
	for _, file := range files {
		fmt.Println("Running migration:", file)

		sqlBytes, err := ioutil.ReadFile(file)
		if err != nil {
			log.Fatal("Error reading file:", file, err)
		}

		// Execute the SQL statements in the file
		_, err = db.Exec(string(sqlBytes))
		if err != nil {
			log.Fatalf("Error running migration %s: %v", file, err)
		}
	}

	fmt.Println("All migrations ran successfully!")

	// Step 5: Verify tables were created
	printTables(db)
}

// printTables prints all tables in the connected database
func printTables(db *sql.DB) {
	fmt.Println("Listing tables in database:")
	rows, err := db.Query("SELECT table_name FROM information_schema.tables WHERE table_schema=DATABASE()")
	if err != nil {
		log.Fatal("Error fetching table list:", err)
	}
	defer rows.Close()

	var tableName string
	for rows.Next() {
		err = rows.Scan(&tableName)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("-", tableName)
	}
}