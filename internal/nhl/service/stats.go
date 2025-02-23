package service

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"

	"api.alexmontague.ca/internal/nhl/models"
	"api.alexmontague.ca/internal/nhl/repository"
)

func GetPlayerStats(gameId int, teamAbbrevs []string) ([]models.PlayerDetail, error) {
	var allPlayers []models.PlayerDetail
	playerChan := make(chan models.PlayerDetail, 50) // Buffered channel to prevent blocking
	errorChan := make(chan error, 2)                 // Buffer for potential errors
	var wg sync.WaitGroup

	// Fetch rosters for both teams concurrently
	for _, abbrev := range teamAbbrevs {
		wg.Add(1)
		go func(teamAbbrev string) {
			defer wg.Done()

			rosterURL := fmt.Sprintf("%s/roster/%s/current", models.NHL_API_BASE, teamAbbrev)
			resp, err := repository.HTTPGetAndCount(rosterURL)
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

					playerURL := fmt.Sprintf("%s/player/%d/landing", models.NHL_API_BASE, p.ID)
					playerResp, err := repository.HTTPGetAndCount(playerURL)
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
					if teamAbbrev == teamAbbrevs[0] {
						playerDetail.OpposingTeamAbbrev = teamAbbrevs[1]
					} else {
						playerDetail.OpposingTeamAbbrev = teamAbbrevs[0]
					}

					select {
					case playerChan <- playerDetail:
					default:
						fmt.Printf("Warning: Could not send player %d to channel\n", p.ID)
					}
				}(player)
			}
			playerWg.Wait()
		}(abbrev)
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

func calculateConfidence(avgShots, avgTOI float64, trend []int, position string) float64 {
	score := 0.0

	// Weight recent shot volume
	score += avgShots * 2

	// Weight recent trend
	if len(trend) >= 3 {
		if trend[2] >= trend[1] && trend[1] >= trend[0] {
			score += 2
		}
	}

	// Ice time correlation
	score += (avgTOI / 20.0) * 1.5

	// Position adjustment
	if position == "D" {
		score *= 0.8
	}

	return math.Round(score*100) / 100
}

// Helper function to get team stats by abbreviation
func getTeamStatsByAbbrev(teamAbbrev string, allTeamStats []models.TeamStats) *models.TeamStats {
	for _, stats := range allTeamStats {
		// Convert full name to abbreviation for comparison
		// This is a simple example - you might need more robust mapping
		if strings.Contains(stats.TeamFullName, teamAbbrev) {
			return &stats
		}
	}
	return nil
}

func calculatePredictedShots(
	playerStats models.PlayerDetail,
	avgShotsLast5 float64,
	seasonShotsPerGame float64,
	teamStats []models.TeamStats,
) float64 {
	currentTeam := getTeamStatsByAbbrev(playerStats.CurrentTeamAbbrev, teamStats)
	opposingTeam := getTeamStatsByAbbrev(playerStats.OpposingTeamAbbrev, teamStats)

	if currentTeam == nil || opposingTeam == nil {
		// Fallback to simple average if team stats aren't available
		return (avgShotsLast5 + seasonShotsPerGame) / 2
	}

	// Base prediction on recent performance
	basePrediction := avgShotsLast5*0.6 + seasonShotsPerGame*0.4

	// Team-based adjustments
	teamOffenseFactor := currentTeam.ShotsForPerGame / 30.0 // 30 shots/game is roughly league average
	teamMatchupFactor := opposingTeam.ShotsAgainstPerGame / 30.0

	// Adjust prediction based on team factors
	adjustedPrediction := basePrediction * teamOffenseFactor * teamMatchupFactor

	// Position-based scaling
	if playerStats.Position == "D" {
		adjustedPrediction *= 0.7 // Defenders typically shoot less
	}

	// Round to 1 decimal place
	return math.Round(adjustedPrediction*10) / 10
}

func CalculateShootingStats(players []models.PlayerDetail, teamStats []models.TeamStats) []models.PlayerStats {
	var stats []models.PlayerStats

	for _, player := range players {
		if len(player.Last5Games) < 5 {
			continue
		}

		// Calculate last 5 games shots
		var totalShots float64
		var shotTrend []int
		for i, game := range player.Last5Games {
			totalShots += float64(game.Shots)
			if i >= 2 {
				shotTrend = append(shotTrend, game.Shots)
			}
		}
		avgShotsLast5 := totalShots / 5

		// Calculate average TOI
		var totalTOI float64
		for _, game := range player.Last5Games {
			timeParts := strings.Split(game.TOI, ":")
			if len(timeParts) == 2 {
				minutes, _ := strconv.Atoi(timeParts[0])
				seconds, _ := strconv.Atoi(timeParts[1])
				totalTOI += float64(minutes) + float64(seconds)/60
			}
		}
		avgTOI := totalTOI / 5

		// Calculate season shots per game
		seasonShotsPerGame := float64(0)
		if player.FeaturedStats.RegularSeason.SubSeason.GamesPlayed > 0 {
			seasonShotsPerGame = float64(player.FeaturedStats.RegularSeason.SubSeason.Shots) /
				float64(player.FeaturedStats.RegularSeason.SubSeason.GamesPlayed)
		}

		// Only include players meeting minimum shot threshold
		if avgShotsLast5 >= models.MIN_SHOTS {
			predictedShots := calculatePredictedShots(player, avgShotsLast5, seasonShotsPerGame, teamStats)

			stats = append(stats, models.PlayerStats{
				PlayerId:           player.PlayerId,
				Name:               fmt.Sprintf("%s %s", player.FirstName.Default, player.LastName.Default),
				Position:           player.Position,
				Team:               player.CurrentTeamAbbrev,
				AvgShotsLast5:      avgShotsLast5,
				ShotTrend:          shotTrend,
				AvgTOI:             avgTOI,
				SeasonShotsPerGame: seasonShotsPerGame,
				PredictedGameShots: predictedShots,
				Confidence:         calculateConfidence(avgShotsLast5, avgTOI, shotTrend, player.Position),
			})
		}
	}

	// Sort by confidence
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Confidence > stats[j].Confidence
	})

	return stats
}
