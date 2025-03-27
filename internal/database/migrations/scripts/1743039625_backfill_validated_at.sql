-- Migration: backfill_validated_at
-- Created at: 2024-09-23T12:00:00Z

-- Write your UP migration here
UPDATE game_predictions
SET validated_at = '2025-03-01 00:00:00'
WHERE validated_at IS NULL;

-- DOWN
-- No down migration needed as we can't restore NULL values without additional data
-- This migration is not reversible