-- Migration: add_validated_at
-- Created at: 2025-03-15T12:46:33-04:00

-- Write your UP migration here
ALTER TABLE game_predictions ADD COLUMN validated_at TIMESTAMP;

-- DOWN
ALTER TABLE game_predictions DROP COLUMN validated_at;
