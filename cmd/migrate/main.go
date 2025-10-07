package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"api.alexmontague.ca/internal/database"
	"api.alexmontague.ca/internal/database/migrations"
)

func main() {
	// Define command-line flags
	upCmd := flag.Bool("up", false, "Apply pending migrations")
	downCmd := flag.Bool("down", false, "Roll back the last migration")
	createCmd := flag.String("create", "", "Create a new migration with the given name")
	migrationsDir := flag.String("dir", "./internal/database/migrations/scripts", "Directory containing migration files")
	dbPath := flag.String("db", database.DB_PATH, "Path to SQLite database file")

	flag.Parse()

	// Validate commands
	cmdCount := 0
	if *upCmd {
		cmdCount++
	}
	if *downCmd {
		cmdCount++
	}
	if *createCmd != "" {
		cmdCount++
	}

	if cmdCount != 1 {
		fmt.Println("Please specify exactly one command: -up, -down, or -create NAME")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Create migration
	if *createCmd != "" {
		path, err := migrations.CreateMigration(*migrationsDir, *createCmd)
		if err != nil {
			log.Fatalf("Failed to create migration: %v", err)
		}
		fmt.Printf("Created migration file: %s\n", path)
		return
	}

	// Initialize database
	if err := database.InitDB(*dbPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Run migrations
	if *upCmd {
		if err := migrations.MigrateUp(database.DB, *migrationsDir); err != nil {
			log.Fatalf("Failed to apply migrations: %v", err)
		}
	} else if *downCmd {
		if err := migrations.MigrateDown(database.DB, *migrationsDir); err != nil {
			log.Fatalf("Failed to roll back migration: %v", err)
		}
	}
}
