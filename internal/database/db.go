package database

import (
	"database/sql"

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

	// Test connection
	if err = DB.Ping(); err != nil {
		return err
	}

	// Create simplified schema with a single table for game predictions
	_, err = DB.Exec(`
	CREATE TABLE IF NOT EXISTS game_predictions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		game_date TEXT NOT NULL,
		game_id INTEGER NOT NULL,
		game_title TEXT NOT NULL,
		away_team_abbrev TEXT NOT NULL,
		away_team_id INTEGER NOT NULL,
		home_team_abbrev TEXT NOT NULL,
		home_team_id INTEGER NOT NULL,
		player_id INTEGER NOT NULL,
		player_name TEXT NOT NULL,
		player_team_abbrev TEXT NOT NULL,
		player_team_id INTEGER NOT NULL,
		predicted_shots REAL NOT NULL,
		confidence REAL NOT NULL,
		actual_shots INTEGER,
		successful BOOLEAN,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(game_id, player_id)
	);

	CREATE INDEX IF NOT EXISTS idx_game_predictions_game_id ON game_predictions(game_id);
	CREATE INDEX IF NOT EXISTS idx_game_predictions_game_date ON game_predictions(game_date);
	CREATE INDEX IF NOT EXISTS idx_game_predictions_player_id ON game_predictions(player_id);
	`)

	return err
}

// Close closes the database connection
func Close() {
	if DB != nil {
		DB.Close()
	}
}
