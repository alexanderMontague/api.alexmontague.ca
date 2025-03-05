package repository

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"api.alexmontague.ca/helpers"
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
				ID           int         `json:"id"`
				AwayTeam     models.Team `json:"awayTeam"`
				HomeTeam     models.Team `json:"homeTeam"`
				Season       int         `json:"season"`
				StartTimeUTC string      `json:"startTimeUTC"`
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
					GameID:       game.ID,
					Title:        fmt.Sprintf("%s @ %s", game.AwayTeam.Abbrev, game.HomeTeam.Abbrev),
					AwayTeam:     game.AwayTeam,
					HomeTeam:     game.HomeTeam,
					Season:       strconv.Itoa(game.Season),
					StartTimeUTC: game.StartTimeUTC,
					EstDate:      week.Date,
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

// map[teamId]restDays
func GetTeamsRest(currGameDay string, games []models.Game) (map[int]int, error) {
	currGameDayTime, err := time.Parse("2006-01-02", currGameDay)
	if err != nil {
		fmt.Println("Error parsing curr game day time:", err)
		return nil, err
	}

	url := fmt.Sprintf("%s/schedule/%s", models.NHL_API_BASE, currGameDayTime.AddDate(0, 0, -6).Format("2006-01-02"))
	resp, err := HTTPGetAndCount(url)
	if err != nil {
		fmt.Println("Error getting teams rest:", err)
		return nil, err
	}
	defer resp.Body.Close()

	var schedule struct {
		GameWeek []struct {
			Date  string `json:"date"`
			Games []struct {
				StartTimeUTC string      `json:"startTimeUTC"`
				AwayTeam     models.Team `json:"awayTeam"`
				HomeTeam     models.Team `json:"homeTeam"`
			} `json:"games"`
		} `json:"gameWeek"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&schedule); err != nil {
		fmt.Println("Error decoding teams rest:", err)
		return nil, err
	}

	// Map to store last game date for each team
	lastGameDate := make(map[int]time.Time)
	// Map to store rest days for each team
	restDays := make(map[int]int)

	// Process games in chronological order
	for i, week := range schedule.GameWeek {
		for _, game := range week.Games {
			currGame := helpers.Find(games, func(g models.Game) bool {
				return g.AwayTeam.Id == game.AwayTeam.Id || g.AwayTeam.Id == game.HomeTeam.Id ||
					g.HomeTeam.Id == game.AwayTeam.Id || g.HomeTeam.Id == game.HomeTeam.Id
			})

			if currGame == nil {
				fmt.Println("currGame not found", game)
				continue
			}

			prevGameTime, err := time.Parse(time.RFC3339, game.StartTimeUTC)
			if err != nil {
				fmt.Println("Error parsing past game time:", err)
				continue
			}

			currGameTime, err := time.Parse(time.RFC3339, currGame.StartTimeUTC)
			if err != nil {
				fmt.Println("Error parsing curr game time:", err)
				continue
			}

			// Skip future games
			if prevGameTime.After(currGameTime) || prevGameTime == currGameTime {
				continue
			}

			// Update last game date for both teams
			lastGameDate[game.HomeTeam.Id] = prevGameTime
			lastGameDate[game.AwayTeam.Id] = prevGameTime

			if i == len(schedule.GameWeek)-1 {
				fmt.Println(lastGameDate)
			}

			// Calculate rest days for teams that haven't played yet
			calculateRestDays(game.HomeTeam.Id, currGameTime, lastGameDate, restDays)
			calculateRestDays(game.AwayTeam.Id, currGameTime, lastGameDate, restDays)
		}
	}

	return restDays, nil
}

// Helpers
func calculateRestDays(teamId int, nextGameTime time.Time, lastGameDate map[int]time.Time, restDays map[int]int) {
	if lastGame, exists := lastGameDate[teamId]; exists {
		// Calculate calendar days between games (ignoring time)
		lastGameDay := lastGame.Truncate(24 * time.Hour)
		nextGameDay := nextGameTime.Truncate(24 * time.Hour)
		days := int(nextGameDay.Sub(lastGameDay).Hours() / 24)

		// Subtract 1 to get rest days (days between games)
		if days > 0 {
			restDays[teamId] = days - 1
		} else {
			restDays[teamId] = 0 // Same day games have 0 rest days
		}
	} else {
		// If no previous game found, set a default high value (e.g., 7 days)
		restDays[teamId] = 7
	}
}
