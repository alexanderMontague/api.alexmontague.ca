package repository

import (
	"encoding/json"
	"fmt"
	"log"

	"api.alexmontague.ca/internal/nhl/models"
)

// FetchActualGameShots retrieves actual shot data from completed games
func FetchActualGameShots(gameIDs []int) (map[int]int, error) {
	// Create map to store player shot results
	results := make(map[int]int)

	// Get unique game IDs
	uniqueGameIDs := make(map[int]bool)
	for _, id := range gameIDs {
		uniqueGameIDs[id] = true
	}

	// Process each unique game
	for gameID := range uniqueGameIDs {
		// NHL API for game stats
		url := fmt.Sprintf("%s/gamecenter/%d/boxscore", models.NHL_API_BASE, gameID)
		resp, err := HTTPGetAndCount(url)
		if err != nil {
			return nil, err
		}

		type PlayerShotResult struct {
			PlayerID int `json:"playerId"`
			Sog      int `json:"sog"`
		}

		var boxscore struct {
			PlayerByGameStats struct {
				AwayTeam struct {
					Forwards []PlayerShotResult `json:"forwards"`
					Defense  []PlayerShotResult `json:"defense"`
				} `json:"awayTeam"`
				HomeTeam struct {
					Forwards []PlayerShotResult `json:"forwards"`
					Defense  []PlayerShotResult `json:"defense"`
				} `json:"homeTeam"`
			} `json:"playerByGameStats"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&boxscore); err != nil {
			resp.Body.Close()
			return nil, err
		}
		resp.Body.Close()

		// Process away team
		for _, player := range append(boxscore.PlayerByGameStats.AwayTeam.Forwards, boxscore.PlayerByGameStats.AwayTeam.Defense...) {
			results[player.PlayerID] = player.Sog
		}

		// Process home team
		for _, player := range append(boxscore.PlayerByGameStats.HomeTeam.Forwards, boxscore.PlayerByGameStats.HomeTeam.Defense...) {
			results[player.PlayerID] = player.Sog
		}
	}

	log.Printf("Fetched shots for %d players", len(results))

	return results, nil
}
