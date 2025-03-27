-- Migration: add_model_version_id
-- Created at: 2023-05-09T10:00:00Z

-- Add model_version_id column to game_predictions table
ALTER TABLE game_predictions ADD model_version_id INTEGER DEFAULT 1 NOT NULL;

UPDATE game_predictions
SET model_version_id = 1
WHERE model_version_id IS NULL;

-- DOWN

-- Remove model_version_id column from game_predictions table
-- Note: SQLite doesn't support DROP COLUMN directly, so we would need to recreate the table
-- This is a no-op since we don't want to lose data