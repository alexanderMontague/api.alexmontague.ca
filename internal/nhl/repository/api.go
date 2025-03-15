package repository

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"api.alexmontague.ca/helpers"
	"api.alexmontague.ca/internal/nhl/models"
)

var RequestCount uint64

func httpGetAndCount(url string) (*http.Response, error) {
	atomic.AddUint64(&RequestCount, 1)
	return http.Get(url)
}

func GetPlayerStats(gameId int, teamInfo []models.Team) ([]models.PlayerDetail, error) {
	var allPlayers []models.PlayerDetail
	playerChan := make(chan models.PlayerDetail, 50) // Buffered channel to prevent blocking
	errorChan := make(chan error, 2)                 // Buffer for potential errors
	var wg sync.WaitGroup

	// Fetch rosters for both teams concurrently
	for _, team := range teamInfo {
		wg.Add(1)
		go func(team models.Team) {
			defer wg.Done()

			rosterURL := fmt.Sprintf("%s/roster/%s/current", models.NHL_API_BASE, team.Abbrev)
			resp, err := httpGetAndCount(rosterURL)
			if err != nil {
				errorChan <- err
				return
			}
			defer resp.Body.Close()

			var roster models.RosterResponse
			if err := json.NewDecoder(resp.Body).Decode(&roster); err != nil {
				errorChan <- err
				return
			}

			// Combine forwards and defensemen
			allSkaters := append(roster.Forwards, roster.Defensemen...)

			// Create a WaitGroup for players within this team
			var playerWg sync.WaitGroup
			for _, player := range allSkaters {
				playerWg.Add(1)
				go func(p models.Player) {
					defer playerWg.Done()

					playerURL := fmt.Sprintf("%s/player/%d/landing", models.NHL_API_BASE, p.Id)
					playerResp, err := httpGetAndCount(playerURL)
					if err != nil {
						errorChan <- err
						return
					}
					defer playerResp.Body.Close()

					var playerDetail models.PlayerDetail
					if err := json.NewDecoder(playerResp.Body).Decode(&playerDetail); err != nil {
						errorChan <- err
						return
					}
					if team.Id == teamInfo[0].Id {
						playerDetail.OpposingTeamId = teamInfo[1].Id
						playerDetail.OpposingTeamAbbrev = teamInfo[1].Abbrev
					} else {
						playerDetail.OpposingTeamId = teamInfo[0].Id
						playerDetail.OpposingTeamAbbrev = teamInfo[0].Abbrev
					}

					select {
					case playerChan <- playerDetail:
					default:
						fmt.Printf("Warning: Could not send player %d to channel\n", p.Id)
					}
				}(player)
			}
			playerWg.Wait()
		}(team)
	}

	// Wait in a separate goroutine and close channels when done
	go func() {
		wg.Wait()
		close(playerChan)
		close(errorChan)
	}()

	// Check for errors first
	select {
	case err := <-errorChan:
		if err != nil {
			return nil, err
		}
	default:
	}

	// Collect all players from channel
	for player := range playerChan {
		allPlayers = append(allPlayers, player)
	}

	return allPlayers, nil
}

func GetUpcomingGames(date string) ([]models.Game, error) {
	url := fmt.Sprintf("%s/schedule/%s", models.NHL_API_BASE, date)
	resp, err := httpGetAndCount(url)
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
	resp, err := httpGetAndCount(url)
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
	resp, err := httpGetAndCount(url)
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
		resp, err := httpGetAndCount(url)
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
