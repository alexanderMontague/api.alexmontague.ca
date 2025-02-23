package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"api.alexmontague.ca/helpers"
	"api.alexmontague.ca/internal/nhl/models"
	"api.alexmontague.ca/internal/nhl/repository"
	"api.alexmontague.ca/internal/nhl/service"
)

// Route : '/nhl/shots?date=2025-02-23
// Type  : 'GET'
func GetPlayerShotStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	date := r.URL.Query().Get("date")
	if date == "" {
		now := time.Now()
		loc, _ := time.LoadLocation("America/New_York")
		date = now.In(loc).Format("2006-01-02")
	}

	fmt.Println("Fetching upcoming games for date:", date)

	games, err := repository.GetUpcomingGames(date)
	if err != nil {
		json.NewEncoder(w).Encode(helpers.Response{
			Error:   true,
			Code:    500,
			Message: "Error fetching NHL games",
		})
		return
	}

	if len(games) == 0 {
		json.NewEncoder(w).Encode(helpers.Response{
			Error:   true,
			Code:    404,
			Message: "No NHL games found for date",
		})
		return
	}

	teamStats, err := repository.GetAllTeamStats(games[0].Season)
	if err != nil {
		json.NewEncoder(w).Encode(helpers.Response{
			Error:   true,
			Code:    500,
			Message: "Error fetching NHL team stats",
		})
		return
	}

	var allPlayers []models.PlayerStats
	for _, game := range games {
		players, err := service.GetPlayerStats(game.GameID, []models.Team{game.AwayTeam, game.HomeTeam})
		if err != nil {
			fmt.Println("Error fetching player stats:", err)
			continue
		}
		playerStats := service.CalculateShootingStats(players, teamStats)
		allPlayers = append(allPlayers, playerStats...)
	}

	var gamesWithPlayers []struct {
		models.Game
		Players []models.PlayerStats `json:"players"`
	}

	for _, game := range games {
		gamesWithPlayers = append(gamesWithPlayers, struct {
			models.Game
			Players []models.PlayerStats `json:"players"`
		}{
			Game: game,
			Players: helpers.Filter(allPlayers, func(player models.PlayerStats) bool {
				return player.TeamId == game.AwayTeam.Id || player.TeamId == game.HomeTeam.Id
			}),
		})
	}

	fmt.Printf("Total API requests made: %d\n", atomic.LoadUint64(&repository.RequestCount))

	json.NewEncoder(w).Encode(gamesWithPlayers)
}
