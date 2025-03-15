package database

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	"api.alexmontague.ca/internal/database/migrations"
	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// InitDB initializes the database connection and runs migrations
func InitDB(dbPath string) error {
	// Ensure directory exists
	dbDir := filepath.Dir(dbPath)
	if _, err := os.Stat(dbDir); os.IsNotExist(err) {
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			return err
		}
	}

	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	// Test connection
	if err = DB.Ping(); err != nil {
		return err
	}

	// Run migrations instead of direct schema creation
	migrationsDir := "./internal/database/migrations/scripts"
	log.Println("Running database migrations...")
	if err := migrations.MigrateUp(DB, migrationsDir); err != nil {
		log.Printf("Error running migrations: %v", err)
		return err
	}
	log.Println("Database migrations completed successfully")

	return nil
}

// Close closes the database connection
func Close() {
	if DB != nil {
		DB.Close()
	}
}
