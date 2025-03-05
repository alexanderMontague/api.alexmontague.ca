package repository

import (
	"fmt"
	"time"

	"api.alexmontague.ca/internal/database"
)

// StoreActualShots saves the actual shot results
func StoreActualShots(gameID int, playerID int, actualShots int) error {
	_, err := database.DB.Exec(`
	INSERT OR REPLACE INTO shot_results
	(game_id, player_id, actual_shots, result_time)
	VALUES (?, ?, ?, ?)`,
		gameID,
		playerID,
		actualShots,
		time.Now(),
	)
	return err
}

// GetAccuracyStats retrieves prediction accuracy metrics
type PredictionAccuracy struct {
	GameID         int
	Date           string
	PlayerID       int
	PlayerName     string
	PredictedShots float64
	ActualShots    int
	Difference     float64
	Confidence     float64
}

// GetPredictionAccuracy gets accuracy data for analysis
func GetPredictionAccuracy(days int) ([]PredictionAccuracy, error) {
	query := `
	SELECT g.game_id, g.date, p.player_id, p.player_name,
	       p.predicted_shots, r.actual_shots,
	       ABS(p.predicted_shots - r.actual_shots) as difference,
	       p.confidence
	FROM games g
	JOIN player_predictions p ON g.game_id = p.game_id
	JOIN shot_results r ON p.game_id = r.game_id AND p.player_id = r.player_id
	WHERE g.date >= date('now', ?) AND g.date <= date('now')
	ORDER BY g.date DESC, difference ASC`

	rows, err := database.DB.Query(query, fmt.Sprintf("-%d days", days))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []PredictionAccuracy
	for rows.Next() {
		var r PredictionAccuracy
		if err := rows.Scan(
			&r.GameID,
			&r.Date,
			&r.PlayerID,
			&r.PlayerName,
			&r.PredictedShots,
			&r.ActualShots,
			&r.Difference,
			&r.Confidence,
		); err != nil {
			return nil, err
		}
		results = append(results, r)
	}

	return results, nil
}
