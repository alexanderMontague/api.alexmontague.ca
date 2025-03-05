package database

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// InitDB initializes the database connection
func InitDB(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	// Set connection limits
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(5 * time.Minute)

	// Verify connection
	if err = DB.Ping(); err != nil {
		return err
	}

	// Create tables if they don't exist
	if err = createTables(); err != nil {
		return err
	}

	return nil
}

// createTables sets up the database schema
func createTables() error {
	// Games table
	_, err := DB.Exec(`
	CREATE TABLE IF NOT EXISTS games (
		game_id INTEGER PRIMARY KEY,
		date TEXT NOT NULL,
		away_team_id INTEGER NOT NULL,
		home_team_id INTEGER NOT NULL,
		season TEXT NOT NULL,
		start_time TEXT NOT NULL,
		status TEXT DEFAULT 'scheduled',
		processed BOOLEAN DEFAULT 0
	)`)
	if err != nil {
		return err
	}

	// Player predictions table
	_, err = DB.Exec(`
	CREATE TABLE IF NOT EXISTS player_predictions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		game_id INTEGER NOT NULL,
		player_id INTEGER NOT NULL,
		player_name TEXT NOT NULL,
		team_id INTEGER NOT NULL,
		team_abbrev TEXT NOT NULL,
		position TEXT NOT NULL,
		predicted_shots REAL NOT NULL,
		confidence REAL NOT NULL,
		avg_shots_last5 REAL NOT NULL,
		rest_days INTEGER,
		prediction_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (game_id) REFERENCES games(game_id),
		UNIQUE(game_id, player_id)
	)`)
	if err != nil {
		return err
	}

	// Results table for validation
	_, err = DB.Exec(`
	CREATE TABLE IF NOT EXISTS shot_results (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		game_id INTEGER NOT NULL,
		player_id INTEGER NOT NULL,
		actual_shots INTEGER NOT NULL,
		result_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (game_id) REFERENCES games(game_id),
		UNIQUE(game_id, player_id)
	)`)
	if err != nil {
		return err
	}

	// Create indexes for better performance
	_, err = DB.Exec(`
	CREATE INDEX IF NOT EXISTS idx_predictions_game_id ON player_predictions(game_id);
	CREATE INDEX IF NOT EXISTS idx_results_game_id ON shot_results(game_id);
	CREATE INDEX IF NOT EXISTS idx_games_date ON games(date);
	CREATE INDEX IF NOT EXISTS idx_games_status ON games(status);
	`)

	return err
}

// Close closes the database connection
func Close() {
	if DB != nil {
		DB.Close()
	}
}
