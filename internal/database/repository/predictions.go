package repository

import (
	"time"

	"api.alexmontague.ca/internal/database"
	"api.alexmontague.ca/internal/nhl/models"
)

// StorePredictions saves player shot predictions to the database
func StorePredictions(gameID int, predictions []models.PlayerStats) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
	INSERT OR REPLACE INTO player_predictions
	(game_id, player_id, player_name, team_id, team_abbrev, position,
	 predicted_shots, confidence, avg_shots_last5, rest_days, prediction_time)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, p := range predictions {
		_, err = stmt.Exec(
			gameID,
			p.PlayerId,
			p.Name,
			p.TeamId,
			p.TeamAbbrev,
			p.Position,
			p.PredictedGameShots,
			p.Confidence,
			p.AvgShotsLast5,
			p.RestDays,
			time.Now(),
		)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

// GetPredictionsForGame retrieves all predictions for a specific game
func GetPredictionsForGame(gameID int) ([]models.PlayerStats, error) {
	rows, err := database.DB.Query(`
	SELECT player_id, player_name, team_id, team_abbrev, position,
	       predicted_shots, confidence, avg_shots_last5, rest_days
	FROM player_predictions
	WHERE game_id = ?
	ORDER BY confidence DESC`, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var predictions []models.PlayerStats
	for rows.Next() {
		var p models.PlayerStats
		if err := rows.Scan(
			&p.PlayerId,
			&p.Name,
			&p.TeamId,
			&p.TeamAbbrev,
			&p.Position,
			&p.PredictedGameShots,
			&p.Confidence,
			&p.AvgShotsLast5,
			&p.RestDays,
		); err != nil {
			return nil, err
		}
		predictions = append(predictions, p)
	}

	return predictions, nil
}
