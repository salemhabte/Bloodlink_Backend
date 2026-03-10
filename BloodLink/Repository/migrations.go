package Repository

import (
	"log"
	"os"
	"path/filepath"
)

func RunMigrations() {
	migrationFolder := "migrations"

	// Get all .sql files in order
	files, err := filepath.Glob(filepath.Join(migrationFolder, "*.sql"))
	if err != nil {
		log.Fatalf("Error reading migration folder: %v", err)
	}
	if len(files) == 0 {
		log.Printf("No migration files found in: %s", migrationFolder)
		return
	}

	log.Printf("Found %d migration files. Starting migration...", len(files))

	for _, file := range files {
		log.Printf("Running migration: %s", file)

		sqlBytes, err := os.ReadFile(file)
		if err != nil {
			log.Fatalf("Error reading migration file %s: %v", file, err)
		}

		// Execute the SQL statements in the file
		_, err = DB.Exec(string(sqlBytes))
		if err != nil {
			log.Fatalf("Migration failed on file %s: %v", file, err)
		}
	}

	log.Println("All migrations completed successfully!")
}
