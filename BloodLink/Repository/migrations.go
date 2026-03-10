package Repository

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func RunMigrations() {
	migrationFolder := "migrations"

	// 1. Ensure schema_migrations table exists
	createTableSQL := `CREATE TABLE IF NOT EXISTS schema_migrations (
		migration_file VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`
	if _, err := DB.Exec(createTableSQL); err != nil {
		log.Fatalf("Error creating schema_migrations table: %v", err)
	}

	// 2. Normalize Table Casing (Fix for Linux/Render case-sensitivity)
	normalizeTableCasing()

	// 3. Get all .sql files in order
	files, err := filepath.Glob(filepath.Join(migrationFolder, "*.sql"))
	if err != nil {
		log.Fatalf("Error reading migration folder: %v", err)
	}
	if len(files) == 0 {
		log.Printf("No migration files found in: %s", migrationFolder)
		return
	}

	log.Printf("Found %d migration files. Checking for pending migrations...", len(files))

	for _, file := range files {
		fileName := filepath.Base(file)

		// 4. Check if migration was already applied
		var count int
		err := DB.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE migration_file = ?", fileName).Scan(&count)
		if err != nil {
			log.Fatalf("Error checking migration status for %s: %v", fileName, err)
		}

		if count > 0 {
			continue // Already applied
		}

		log.Printf("Applying migration: %s", fileName)

		sqlBytes, err := os.ReadFile(file)
		if err != nil {
			log.Fatalf("Error reading migration file %s: %v", file, err)
		}

		// 5. Split and Execute the SQL statements
		statements := strings.Split(string(sqlBytes), ";")
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if _, err := DB.Exec(stmt); err != nil {
				// Handle MySQL "Idempotency" errors (already exists)
				// 1060: Duplicate column name
				// 1061: Duplicate key name
				// 1050: Table already exists
				// 1091: Can't DROP 'column'; check that it exists
				errStr := err.Error()
				if strings.Contains(errStr, "Error 1060") ||
					strings.Contains(errStr, "Error 1061") ||
					strings.Contains(errStr, "Error 1050") ||
					strings.Contains(errStr, "Error 1091") {
					log.Printf("[MIGRATION WARNING] Skipping statement in %s: %v", fileName, err)
					continue
				}
				log.Fatalf("Migration failed on file %s at statement [%s]: %v", fileName, stmt, err)
			}
		}

		_, err = DB.Exec("INSERT INTO schema_migrations (migration_file) VALUES (?)", fileName)
		if err != nil {
			log.Fatalf("Error recording migration status for %s: %v", fileName, err)
		}
	}

	log.Println("Database is up to date!")
}

func normalizeTableCasing() {
	rows, err := DB.Query("SHOW TABLES")
	if err != nil {
		log.Printf("[WARNING] Could not list tables for normalization: %v", err)
		return
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err == nil {
			tables = append(tables, table)
		}
	}

	for _, table := range tables {
		lowerTable := strings.ToLower(table)
		if table != lowerTable {
			log.Printf("Normalizing table casing: %s -> %s", table, lowerTable)
			// Use backticks to handle special characters or keywords
			renameSQL := fmt.Sprintf("RENAME TABLE `%s` TO `%s` ", table, lowerTable)
			if _, err := DB.Exec(renameSQL); err != nil {
				log.Printf("[WARNING] Failed to rename table %s to %s: %v", table, lowerTable, err)
			}
		}
	}
}
