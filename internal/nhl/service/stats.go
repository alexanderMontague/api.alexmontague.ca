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

					playerURL := fmt.Sprintf("%s/player/%d/landing", models.NHL_API_BASE, p.Id)
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

func calculateConfidence(avgShots, avgTOI float64, trend []int, position string) float64 {
	score := 0.0

	// Shot volume weight (0-5 points)
	score += math.Min(avgShots/2, 5.0)

	// Recent trend analysis (0-3 points)
	if len(trend) >= 3 {
		// Increasing trend
		if trend[2] > trend[1] && trend[1] > trend[0] {
			score += 3
		} else if trend[2] > trend[0] { // Generally improving
			score += 1.5
		}

		// Consistency factor
		variance := calculateVariance(trend)
		score += math.Max(0, 2-variance) // Lower variance = higher score
	}

	// Ice time weight (0-4 points)
	score += (avgTOI / 20.0) * 4

	// Position adjustment
	if position == "D" {
		score *= 0.85
	}

	// Normalize to 0-100 scale
	normalizedScore := (score / 12.0) * 100

	return math.Round(math.Min(normalizedScore, 100)*100) / 100
}

func getLeagueShotAverage(allTeamStats []models.TeamStats) float64 {
	var totalShots float64
	for _, stats := range allTeamStats {
		totalShots += stats.ShotsForPerGame
	}
	return totalShots / float64(len(allTeamStats))
}

func getTeamStatsById(teamId int, allTeamStats []models.TeamStats) *models.TeamStats {
	for _, stats := range allTeamStats {
		if stats.TeamId == teamId {
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
	currentTeam := getTeamStatsById(playerStats.CurrentTeamId, teamStats)
	opposingTeam := getTeamStatsById(playerStats.OpposingTeamId, teamStats)

	if currentTeam == nil || opposingTeam == nil {
		return (avgShotsLast5 + seasonShotsPerGame) / 2
	}

	// Weight recent performance more heavily
	basePrediction := avgShotsLast5*0.7 + seasonShotsPerGame*0.3

	// Team pace factors
	leagueShotAverage := getLeagueShotAverage(teamStats)
	gamePaceFactor := (currentTeam.ShotsForPerGame + opposingTeam.ShotsAgainstPerGame) / (2 * leagueShotAverage)

	// Team strength adjustments
	teamOffenseFactor := math.Pow(currentTeam.ShotsForPerGame/leagueShotAverage, 0.8)
	teamDefenseFactor := math.Pow(opposingTeam.ShotsAgainstPerGame/leagueShotAverage, 0.6)

	// Position-based adjustment
	positionFactor := 1.0
	if playerStats.Position == "D" {
		positionFactor = 0.75
	}

	// Calculate ice time factor (assuming 20 mins is max TOI)
	avgTOIMinutes := calculateAvgTOIMinutes(playerStats.Last5Games)
	icetimeFactor := math.Min(avgTOIMinutes/20.0, 1.0)

	// Combine all factors
	adjustedPrediction := basePrediction *
		gamePaceFactor *
		teamOffenseFactor *
		teamDefenseFactor *
		positionFactor *
		icetimeFactor

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
		var shotsLast5 []int
		var shotTrend []int
		for i, game := range player.Last5Games {
			totalShots += float64(game.Shots)
			shotsLast5 = append(shotsLast5, game.Shots)
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
			stats = append(stats, models.PlayerStats{
				PlayerId:           player.PlayerId,
				Name:               fmt.Sprintf("%s %s", player.FirstName.Default, player.LastName.Default),
				Position:           player.Position,
				TeamAbbrev:         player.CurrentTeamAbbrev,
				TeamId:             player.CurrentTeamId,
				ShotsLast5:         shotsLast5,
				AvgShotsLast5:      avgShotsLast5,
				ShotTrend:          shotTrend,
				AvgTOI:             avgTOI,
				SeasonShotsPerGame: seasonShotsPerGame,
				PredictedGameShots: calculatePredictedShots(player, avgShotsLast5, seasonShotsPerGame, teamStats),
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

// Helper functions
func calculateAvgTOIMinutes(games []models.Last5Game) float64 {
	var totalMinutes float64
	for _, game := range games {
		minutes := parseTimeOnIce(game.TOI)
		totalMinutes += minutes
	}
	return totalMinutes / float64(len(games))
}

func parseTimeOnIce(toi string) float64 {
	parts := strings.Split(toi, ":")
	if len(parts) != 2 {
		return 0
	}
	minutes, _ := strconv.ParseFloat(parts[0], 64)
	seconds, _ := strconv.ParseFloat(parts[1], 64)
	return minutes + seconds/60
}

func calculateVariance(numbers []int) float64 {
	var sum float64
	var mean float64

	// Calculate mean
	for _, num := range numbers {
		sum += float64(num)
	}
	mean = sum / float64(len(numbers))

	// Calculate variance
	var variance float64
	for _, num := range numbers {
		diff := float64(num) - mean
		variance += diff * diff
	}
	variance = variance / float64(len(numbers))

	return variance
}
