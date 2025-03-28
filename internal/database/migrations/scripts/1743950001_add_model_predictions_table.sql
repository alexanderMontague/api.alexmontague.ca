-- Migration: add_model_predictions_table
-- Created at: 2023-05-09T12:00:00Z

-- Create model_predictions table for tracking multiple model predictions in parallel
CREATE TABLE model_predictions
(
    id INTEGER PRIMARY KEY,
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
    model_version_id INTEGER NOT NULL,
    predicted_shots REAL NOT NULL,
    confidence REAL NOT NULL,
    actual_shots INTEGER,
    successful INTEGER,
    created_at TEXT NOT NULL,
    validated_at TEXT,
    UNIQUE(game_id, player_id, model_version_id)
);

-- Create indexes for efficient queries
CREATE INDEX idx_model_predictions_game_id ON model_predictions(game_id);
CREATE INDEX idx_model_predictions_player_id ON model_predictions(player_id);
CREATE INDEX idx_model_predictions_model_version_id ON model_predictions(model_version_id);
CREATE INDEX idx_model_predictions_game_date ON model_predictions(game_date);

-- DOWN

-- Drop model_predictions table
DROP TABLE IF EXISTS model_predictions;