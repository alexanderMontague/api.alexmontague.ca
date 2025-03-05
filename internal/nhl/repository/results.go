package repository

import (
	"encoding/json"
	"fmt"

	"api.alexmontague.ca/internal/nhl/models"
)

// FetchActualGameShots retrieves actual shot data from completed games
func FetchActualGameShots(gameID int) (map[int]int, error) {
	// NHL API for game stats
	url := fmt.Sprintf("%s/gamecenter/%d/boxscore", models.NHL_API_BASE, gameID)
	resp, err := HTTPGetAndCount(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var boxscore struct {
		AwayTeam struct {
			Players map[string]struct {
				PlayerID int `json:"playerId"`
				Stats    struct {
					Shots int `json:"shots"`
				} `json:"stats"`
			} `json:"players"`
		} `json:"awayTeam"`
		HomeTeam struct {
			Players map[string]struct {
				PlayerID int `json:"playerId"`
				Stats    struct {
					Shots int `json:"shots"`
				} `json:"stats"`
			} `json:"players"`
		} `json:"homeTeam"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&boxscore); err != nil {
		return nil, err
	}

	// Combine results from both teams
	results := make(map[int]int)

	// Process away team
	for _, player := range boxscore.AwayTeam.Players {
		results[player.PlayerID] = player.Stats.Shots
	}

	// Process home team
	for _, player := range boxscore.HomeTeam.Players {
		results[player.PlayerID] = player.Stats.Shots
	}

	return results, nil
}
