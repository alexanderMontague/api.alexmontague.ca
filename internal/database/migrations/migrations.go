package migrations

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Migration represents a single database migration
type Migration struct {
	Version   int64
	Name      string
	UpSQL     string
	DownSQL   string
	AppliedAt *time.Time
}

// CreateMigrationsTable ensures the migrations tracking table exists
func CreateMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`)
	return err
}

// GetAppliedMigrations returns all migrations that have been applied
func GetAppliedMigrations(db *sql.DB) (map[int64]time.Time, error) {
	rows, err := db.Query("SELECT version, applied_at FROM schema_migrations ORDER BY version")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[int64]time.Time)
	for rows.Next() {
		var version int64
		var appliedAt time.Time
		if err := rows.Scan(&version, &appliedAt); err != nil {
			return nil, err
		}
		applied[version] = appliedAt
	}
	return applied, rows.Err()
}

// LoadMigrations loads all migration files from the specified directory
func LoadMigrations(migrationsDir string) ([]Migration, error) {
	var migrations []Migration

	err := filepath.WalkDir(migrationsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".sql") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		filename := filepath.Base(path)
		parts := strings.Split(filename, "_")
		if len(parts) < 2 {
			return fmt.Errorf("invalid migration filename: %s", filename)
		}

		version, err := parseVersion(parts[0])
		if err != nil {
			return err
		}

		name := strings.Join(parts[1:], "_")
		name = strings.TrimSuffix(name, ".sql")

		sections := strings.Split(string(content), "-- DOWN")
		if len(sections) != 2 {
			return fmt.Errorf("migration %s must contain '-- DOWN' separator", filename)
		}

		upSQL := strings.TrimSpace(sections[0])
		downSQL := strings.TrimSpace(sections[1])

		migrations = append(migrations, Migration{
			Version: version,
			Name:    name,
			UpSQL:   upSQL,
			DownSQL: downSQL,
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// MigrateUp applies all pending migrations
func MigrateUp(db *sql.DB, migrationsDir string) error {
	if err := CreateMigrationsTable(db); err != nil {
		return err
	}

	applied, err := GetAppliedMigrations(db)
	if err != nil {
		return err
	}

	migrations, err := LoadMigrations(migrationsDir)
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		if _, ok := applied[migration.Version]; ok {
			// Migration already applied
			continue
		}

		tx, err := db.Begin()
		if err != nil {
			return err
		}

		log.Printf("Applying migration %d: %s", migration.Version, migration.Name)

		if _, err := tx.Exec(migration.UpSQL); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to apply migration %d: %w", migration.Version, err)
		}

		if _, err := tx.Exec(
			"INSERT INTO schema_migrations (version, name) VALUES (?, ?)",
			migration.Version, migration.Name,
		); err != nil {
			tx.Rollback()
			return err
		}

		if err := tx.Commit(); err != nil {
			return err
		}

		log.Printf("Successfully applied migration %d", migration.Version)
	}

	return nil
}

// MigrateDown rolls back the last migration
func MigrateDown(db *sql.DB, migrationsDir string) error {
	if err := CreateMigrationsTable(db); err != nil {
		return err
	}

	var lastVersion int64
	var lastMigrationName string

	err := db.QueryRow(
		"SELECT version, name FROM schema_migrations ORDER BY version DESC LIMIT 1",
	).Scan(&lastVersion, &lastMigrationName)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("No migrations to roll back")
			return nil
		}
		return err
	}

	migrations, err := LoadMigrations(migrationsDir)
	if err != nil {
		return err
	}

	var lastMigration *Migration
	for i := range migrations {
		if migrations[i].Version == lastVersion {
			lastMigration = &migrations[i]
			break
		}
	}

	if lastMigration == nil {
		return fmt.Errorf("could not find migration with version %d", lastVersion)
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	log.Printf("Rolling back migration %d: %s", lastVersion, lastMigration.Name)

	if _, err := tx.Exec(lastMigration.DownSQL); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to roll back migration %d: %w", lastVersion, err)
	}

	if _, err := tx.Exec("DELETE FROM schema_migrations WHERE version = ?", lastVersion); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.Printf("Successfully rolled back migration %d", lastVersion)
	return nil
}

// CreateMigration creates a new migration file
func CreateMigration(migrationsDir, name string) (string, error) {
	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		return "", err
	}

	version := time.Now().Unix()
	filename := fmt.Sprintf("%d_%s.sql", version, name)
	path := filepath.Join(migrationsDir, filename)

	content := `-- Migration: ` + name + `
-- Created at: ` + time.Now().Format(time.RFC3339) + `

-- Write your UP migration here

-- DOWN

-- Write your DOWN migration here
`

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", err
	}

	return path, nil
}

// Helper function to parse migration version
func parseVersion(s string) (int64, error) {
	var version int64
	_, err := fmt.Sscanf(s, "%d", &version)
	return version, err
}
