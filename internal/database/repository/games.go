package repository

import (
	"time"

	"api.alexmontague.ca/internal/database"
	"api.alexmontague.ca/internal/nhl/models"
)

// InsertGame adds a game to the database
func InsertGame(game models.Game) error {
	_, err := database.DB.Exec(`
	INSERT OR IGNORE INTO games
	(game_id, date, away_team_id, home_team_id, season, start_time, status)
	VALUES (?, ?, ?, ?, ?, ?, ?)`,
		game.GameID,
		time.Now().Format("2006-01-02"),
		game.AwayTeam.Id,
		game.HomeTeam.Id,
		game.Season,
		game.StartTimeUTC,
		"scheduled",
	)
	return err
}

// GetUnprocessedGames returns games that need predictions
func GetUnprocessedGames() ([]models.Game, error) {
	rows, err := database.DB.Query(`
	SELECT game_id, date, away_team_id, home_team_id, season, start_time
	FROM games
	WHERE processed = 0 AND date >= date('now', '-1 day')
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []models.Game
	for rows.Next() {
		var g models.Game
		var awayTeamID, homeTeamID int
		var startTime string

		if err := rows.Scan(&g.GameID, &g.EstDate, &awayTeamID, &homeTeamID, &g.Season, &startTime); err != nil {
			return nil, err
		}

		g.AwayTeam.Id = awayTeamID
		g.HomeTeam.Id = homeTeamID
		g.StartTimeUTC = startTime

		games = append(games, g)
	}

	return games, nil
}

// MarkGameProcessed marks a game as having its predictions processed
func MarkGameProcessed(gameID int) error {
	_, err := database.DB.Exec(`
	UPDATE games SET processed = 1
	WHERE game_id = ?`, gameID)
	return err
}

// GetCompletedUnvalidatedGames returns games that have finished but haven't been validated
func GetCompletedUnvalidatedGames() ([]int, error) {
	rows, err := database.DB.Query(`
	SELECT game_id
	FROM games
	WHERE status = 'final' AND
	      game_id NOT IN (SELECT DISTINCT game_id FROM shot_results)
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var gameIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		gameIDs = append(gameIDs, id)
	}

	return gameIDs, nil
}

// UpdateGameStatus updates a game's status
func UpdateGameStatus(gameID int, status string) error {
	_, err := database.DB.Exec(`
	UPDATE games SET status = ?
	WHERE game_id = ?`, status, gameID)
	return err
}
