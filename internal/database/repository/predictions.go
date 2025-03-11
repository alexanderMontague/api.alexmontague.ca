package repository

import (
	"fmt"

	"api.alexmontague.ca/internal/database"
	"api.alexmontague.ca/internal/nhl/models"
)

// StoreGamePredictions stores predictions for a game in the simplified format
func StoreGamePredictions(game models.GameWithPlayers) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
	INSERT OR REPLACE INTO game_predictions (
		game_date, game_id, game_title,
		away_team_abbrev, away_team_id, home_team_abbrev, home_team_id,
		player_id, player_name, player_team_abbrev, player_team_id,
		predicted_shots, confidence
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)

	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, player := range game.Players {
		_, err = stmt.Exec(
			game.EstDate,
			game.GameID,
			game.Title,
			game.AwayTeam.Abbrev,
			game.AwayTeam.Id,
			game.HomeTeam.Abbrev,
			game.HomeTeam.Id,
			player.PlayerId,
			player.Name,
			player.TeamAbbrev,
			player.TeamId,
			player.PredictedGameShots,
			player.Confidence,
		)

		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

// GetGamePredictionsForDate retrieves all predictions for a specific date
func GetGamePredictionsForDate(date string) ([]models.PredictionRecord, error) {
	query := `
	SELECT id, game_date, game_id, game_title,
	       away_team_abbrev, away_team_id, home_team_abbrev, home_team_id,
	       player_id, player_name, player_team_abbrev, player_team_id,
	       predicted_shots, confidence, actual_shots, successful
	FROM game_predictions
	WHERE game_date = ?;`

	gameRows, err := database.DB.Query(query, date)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer gameRows.Close()

	var predictionRecords []models.PredictionRecord

	for gameRows.Next() {
		var record models.PredictionRecord
		err := gameRows.Scan(
			&record.ID,
			&record.GameDate,
			&record.GameID,
			&record.GameTitle,
			&record.AwayTeamAbbrev,
			&record.AwayTeamID,
			&record.HomeTeamAbbrev,
			&record.HomeTeamID,
			&record.PlayerID,
			&record.PlayerName,
			&record.PlayerTeamAbbrev,
			&record.PlayerTeamID,
			&record.PredictedShots,
			&record.Confidence,
			&record.ActualShots,
			&record.Successful,
		)
		if err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		predictionRecords = append(predictionRecords, record)
	}

	if err := gameRows.Err(); err != nil {
		fmt.Printf("Error: %v\n", err)
		return nil, err
	}

	fmt.Printf("Found %d records\n", len(predictionRecords))
	return predictionRecords, nil
}

func StoreActualShots(predictionRecord models.PredictionRecord, actualShots int) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		UPDATE game_predictions
		SET actual_shots = ?, successful = ?
		WHERE id = ?
	`)

	if err != nil {
		return err
	}
	defer stmt.Close()

	successful := actualShots >= int(predictionRecord.PredictedShots)

	_, err = stmt.Exec(actualShots, successful, predictionRecord.ID)
	if err != nil {
		return err
	}

	return tx.Commit()
}
