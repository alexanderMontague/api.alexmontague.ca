-- Migration: initial_schema
-- Created at: 2024-06-01T00:00:00Z

-- This migration captures the existing schema from database.InitDB
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

-- DOWN

DROP INDEX IF EXISTS idx_game_predictions_player_id;
DROP INDEX IF EXISTS idx_game_predictions_game_date;
DROP INDEX IF EXISTS idx_game_predictions_game_id;
DROP TABLE IF EXISTS game_predictions;