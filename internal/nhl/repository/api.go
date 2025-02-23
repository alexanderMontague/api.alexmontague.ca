package repository

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync/atomic"

	"api.alexmontague.ca/internal/nhl/models"
)

var RequestCount uint64

func HTTPGetAndCount(url string) (*http.Response, error) {
	atomic.AddUint64(&RequestCount, 1)
	return http.Get(url)
}

func GetUpcomingGames(date string) ([]models.Game, error) {
	url := fmt.Sprintf("%s/schedule/%s", models.NHL_API_BASE, date)
	resp, err := HTTPGetAndCount(url)
	if err != nil {
		fmt.Println("Error getting upcoming games:", err)
		return nil, err
	}
	defer resp.Body.Close()

	var schedule struct {
		GameWeek []struct {
			Date  string `json:"date"`
			Games []struct {
				ID       int         `json:"id"`
				AwayTeam models.Team `json:"awayTeam"`
				HomeTeam models.Team `json:"homeTeam"`
				Season   int         `json:"season"`
			} `json:"games"`
		} `json:"gameWeek"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&schedule); err != nil {
		fmt.Println("Error decoding upcoming games:", err)
		return nil, err
	}

	var games []models.Game
	for _, week := range schedule.GameWeek {
		if week.Date == date {
			for _, game := range week.Games {
				games = append(games, models.Game{
					GameID:   game.ID,
					Title:    fmt.Sprintf("%s @ %s", game.AwayTeam.Abbrev, game.HomeTeam.Abbrev),
					AwayTeam: game.AwayTeam,
					HomeTeam: game.HomeTeam,
					Season:   strconv.Itoa(game.Season),
				})
			}
		}
	}

	return games, nil
}

func GetAllTeamStats(season string) ([]models.TeamStats, error) {
	url := fmt.Sprintf("%s/team/summary?cayenneExp=seasonId=%s", models.NHL_STATS_API_BASE, season)
	resp, err := HTTPGetAndCount(url)
	if err != nil {
		fmt.Println("Error getting all team stats:", err)
		return nil, err
	}
	defer resp.Body.Close()

	var teamStatsResponse struct {
		Data []models.TeamStats `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&teamStatsResponse); err != nil {
		fmt.Println("Error decoding team stats:", err)
		return nil, err
	}

	return teamStatsResponse.Data, nil
}

// func GetCurrentSchedule()
