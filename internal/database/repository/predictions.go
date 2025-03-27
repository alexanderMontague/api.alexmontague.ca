package repository

import (
	"fmt"
	"time"

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
		predicted_shots, confidence, created_at, model_version_id
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)

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
			time.Now().Format("2006-01-02 15:04:05"),
			player.ModelVersionID,
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
	       predicted_shots, confidence, actual_shots, successful, created_at, validated_at, model_version_id
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
			&record.CreatedAt,
			&record.ValidatedAt,
			&record.ModelVersionID,
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
		SET actual_shots = ?, successful = ?, validated_at = ?
		WHERE id = ?
	`)

	if err != nil {
		return err
	}
	defer stmt.Close()

	successful := actualShots >= int(predictionRecord.PredictedShots)

	_, err = stmt.Exec(actualShots, successful, time.Now().Format("2006-01-02 15:04:05"), predictionRecord.ID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func GetTotalAccuracy() (float64, error) {
	query := `
	SELECT AVG(successful) FROM game_predictions
		WHERE validated_at IS NOT NULL
		AND validated_at <= datetime('now', '+3 hours');
	`

	row := database.DB.QueryRow(query)
	var accuracy float64
	err := row.Scan(&accuracy)
	if err != nil {
		return 0, err
	}
	return accuracy, nil
}

func GetPlayerPastPredictionAccuracy(playerID int) (float64, error) {
	query := `
	SELECT AVG(successful) FROM game_predictions
	WHERE player_id = ?
	`

	row := database.DB.QueryRow(query, playerID)
	var accuracy float64
	err := row.Scan(&accuracy)
	if err != nil {
		return 0, err
	}
	return accuracy, nil
}

func GetPlayerPredictionRecord(playerID int, gameID *int) (models.PredictionRecord, error) {
	query := `
	SELECT id, game_date, game_id, game_title,
	       away_team_abbrev, away_team_id, home_team_abbrev, home_team_id,
	       player_id, player_name, player_team_abbrev, player_team_id,
	       predicted_shots, confidence, actual_shots, successful, created_at, validated_at, model_version_id
	FROM game_predictions
	WHERE player_id = ?
	`

	args := []interface{}{playerID}

	if gameID != nil {
		query += ` AND game_id = ?`
		args = append(args, *gameID)
	}

	query += ` ORDER BY created_at DESC LIMIT 1`

	row := database.DB.QueryRow(query, args...)
	var predictionRecord models.PredictionRecord
	err := row.Scan(
		&predictionRecord.ID,
		&predictionRecord.GameDate,
		&predictionRecord.GameID,
		&predictionRecord.GameTitle,
		&predictionRecord.AwayTeamAbbrev,
		&predictionRecord.AwayTeamID,
		&predictionRecord.HomeTeamAbbrev,
		&predictionRecord.HomeTeamID,
		&predictionRecord.PlayerID,
		&predictionRecord.PlayerName,
		&predictionRecord.PlayerTeamAbbrev,
		&predictionRecord.PlayerTeamID,
		&predictionRecord.PredictedShots,
		&predictionRecord.Confidence,
		&predictionRecord.ActualShots,
		&predictionRecord.Successful,
		&predictionRecord.CreatedAt,
		&predictionRecord.ValidatedAt,
		&predictionRecord.ModelVersionID,
	)
	if err != nil {
		return models.PredictionRecord{}, err
	}
	return predictionRecord, nil
}

// GetModelAccuracy calculates the accuracy for a specific model version
func GetModelAccuracy(modelVersionID int) (float64, error) {
	query := `
	SELECT AVG(successful) FROM game_predictions
	WHERE validated_at IS NOT NULL
	AND model_version_id = ?
	`

	row := database.DB.QueryRow(query, modelVersionID)
	var accuracy float64
	err := row.Scan(&accuracy)
	if err != nil {
		return 0, err
	}
	return accuracy, nil
}

// GetAllModelsAccuracy returns the accuracy for all model versions
func GetAllModelsAccuracy() (map[int]float64, error) {
	query := `
	SELECT model_version_id, AVG(successful) as accuracy
	FROM game_predictions
	WHERE validated_at IS NOT NULL
	GROUP BY model_version_id
	`

	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make(map[int]float64)
	for rows.Next() {
		var modelID int
		var accuracy float64
		if err := rows.Scan(&modelID, &accuracy); err != nil {
			return nil, err
		}
		results[modelID] = accuracy
	}

	return results, rows.Err()
}
