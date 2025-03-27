package service

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"api.alexmontague.ca/helpers"
	dbRepository "api.alexmontague.ca/internal/database/repository"
	"api.alexmontague.ca/internal/nhl/models"
	"api.alexmontague.ca/internal/nhl/repository"
)

func calculateConfidence(predictedShots, avgTOI float64, trend []int, position string, params models.ModelParameters) float64 {
	var score float64

	// Shot volume (0-4 points)
	// Use predicted shots instead of just average shots
	shotScore := math.Min(predictedShots/4.0, 1.0) * params.ShotScoreMultiplier
	score += shotScore

	// Ice time (0-4 points) - Increased weight
	// 20 mins is typical first line/top pair
	// Additional bonus for high TOI players
	toiBase := math.Min(avgTOI/20.0, 1.0) * params.TOIBaseMultiplier
	toiBonus := 0.0
	if avgTOI > params.TOIBonusThreshold {
		// Extra point for high-minute players
		toiBonus = math.Min((avgTOI-params.TOIBonusThreshold)/4, 1.0) * params.TOIBonusMultiplier
	}
	score += toiBase + toiBonus

	// Trend analysis (0-1.5 points) - Slightly reduced to balance with TOI
	if len(trend) >= 3 {
		trendScore := 0.0

		// Strong upward trend
		if trend[2] > trend[1] && trend[1] > trend[0] {
			trendScore = params.TrendUpwardScore
		} else if trend[2] > trend[0] { // General improvement
			trendScore = params.TrendImprovementScore
		}

		// Penalize high variance
		variance := calculateVariance(trend)
		consistencyFactor := math.Max(0.5, 1.0-variance/4.0)
		trendScore *= consistencyFactor

		score += trendScore
	}

	// Position adjustment
	if position == "D" {
		score *= params.DefenseConfidenceFactor
	}

	// Consistency bonus (0-0.5 points) - Adjusted for balance
	if len(trend) >= 3 {
		variance := calculateVariance(trend)
		consistencyBonus := math.Max(0, 0.5-variance/6.0)
		score += consistencyBonus
	}

	// Normalize to 0-10 scale
	normalizedScore := (score / 10.0) * 10.0

	// Round to 1 decimal place
	return math.Round(normalizedScore*10) / 10
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

// Standard calculation method
func calculatePredictedShotsStandard(
	playerStats models.PlayerDetail,
	avgShotsLast5 float64,
	seasonShotsPerGame float64,
	teamStats []models.TeamStats,
	restDays map[int]int,
	params models.ModelParameters,
) float64 {
	currentTeam := getTeamStatsById(playerStats.CurrentTeamId, teamStats)
	opposingTeam := getTeamStatsById(playerStats.OpposingTeamId, teamStats)

	if currentTeam == nil || opposingTeam == nil {
		return (avgShotsLast5 + seasonShotsPerGame) / 2
	}

	// Weight recent performance more heavily
	basePrediction := avgShotsLast5*params.RecentPerformanceWeight + seasonShotsPerGame*params.SeasonPerformanceWeight

	// Team pace factors
	leagueShotAverage := getLeagueShotAverage(teamStats)
	gamePaceFactor := math.Pow((currentTeam.ShotsForPerGame+opposingTeam.ShotsAgainstPerGame)/(2*leagueShotAverage), params.GamePaceExponent)

	// Team strength adjustments
	teamOffenseFactor := math.Pow(currentTeam.ShotsForPerGame/leagueShotAverage, params.TeamOffenseExponent)
	teamDefenseFactor := math.Pow(opposingTeam.ShotsAgainstPerGame/leagueShotAverage, params.TeamDefenseExponent)

	// Position-based adjustment
	positionFactor := 1.0
	if playerStats.Position == "D" {
		positionFactor = params.DefensePositionFactor
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

	// Rest factor (1.0 is baseline, increases with rest, decreases with back-to-back)
	restFactor := 1.0
	if days, exists := restDays[playerStats.CurrentTeamId]; exists {
		switch {
		case days == 0: // Back-to-back games
			restFactor = params.BackToBackFactor
		case days == 1:
			restFactor = params.OneRestDayFactor
		case days >= 4:
			restFactor = params.FourPlusRestDayFactor
		}
	}

	// Apply rest factor to final prediction
	adjustedPrediction = adjustedPrediction * restFactor

	return math.Round(adjustedPrediction*10) / 10
}

// Weighted recency calculation method that uses individual game weighting
func calculatePredictedShotsWeightedRecency(
	playerStats models.PlayerDetail,
	shotsLast5 []int,
	seasonShotsPerGame float64,
	teamStats []models.TeamStats,
	restDays map[int]int,
	params models.ModelParameters,
) float64 {
	if len(shotsLast5) < 5 {
		return calculatePredictedShotsStandard(playerStats, 0, seasonShotsPerGame, teamStats, restDays, params)
	}

	// Apply specific weights to each of the last 5 games
	weightedRecentPerformance := float64(0)
	weightedRecentPerformance += float64(shotsLast5[0]) * params.LastGameWeight
	weightedRecentPerformance += float64(shotsLast5[1]) * params.SecondLastGameWeight
	weightedRecentPerformance += float64(shotsLast5[2]) * params.ThirdLastGameWeight
	weightedRecentPerformance += float64(shotsLast5[3]) * params.FourthLastGameWeight
	weightedRecentPerformance += float64(shotsLast5[4]) * params.FifthLastGameWeight

	// Use weighted recent performance but blend with standard calculation
	currentTeam := getTeamStatsById(playerStats.CurrentTeamId, teamStats)
	opposingTeam := getTeamStatsById(playerStats.OpposingTeamId, teamStats)

	if currentTeam == nil || opposingTeam == nil {
		return (weightedRecentPerformance + seasonShotsPerGame) / 2
	}

	// Weight recent weighted performance more heavily than season stats
	basePrediction := weightedRecentPerformance*params.RecentPerformanceWeight + seasonShotsPerGame*params.SeasonPerformanceWeight

	// Team pace factors
	leagueShotAverage := getLeagueShotAverage(teamStats)
	gamePaceFactor := math.Pow((currentTeam.ShotsForPerGame+opposingTeam.ShotsAgainstPerGame)/(2*leagueShotAverage), params.GamePaceExponent)

	// Team strength adjustments
	teamOffenseFactor := math.Pow(currentTeam.ShotsForPerGame/leagueShotAverage, params.TeamOffenseExponent)
	teamDefenseFactor := math.Pow(opposingTeam.ShotsAgainstPerGame/leagueShotAverage, params.TeamDefenseExponent)

	// Position-based adjustment
	positionFactor := 1.0
	if playerStats.Position == "D" {
		positionFactor = params.DefensePositionFactor
	}

	// Calculate ice time factor
	avgTOIMinutes := calculateAvgTOIMinutes(playerStats.Last5Games)
	icetimeFactor := math.Min(avgTOIMinutes/20.0, 1.0)

	// Combine all factors
	adjustedPrediction := basePrediction *
		gamePaceFactor *
		teamOffenseFactor *
		teamDefenseFactor *
		positionFactor *
		icetimeFactor

	// Rest factor
	restFactor := 1.0
	if days, exists := restDays[playerStats.CurrentTeamId]; exists {
		switch {
		case days == 0:
			restFactor = params.BackToBackFactor
		case days == 1:
			restFactor = params.OneRestDayFactor
		case days >= 4:
			restFactor = params.FourPlusRestDayFactor
		}
	}

	// Apply rest factor
	adjustedPrediction = adjustedPrediction * restFactor

	return math.Round(adjustedPrediction*10) / 10
}

// TOI-driven calculation method that prioritizes ice time
func calculatePredictedShotsTOIDriven(
	playerStats models.PlayerDetail,
	avgShotsLast5 float64,
	seasonShotsPerGame float64,
	teamStats []models.TeamStats,
	restDays map[int]int,
	params models.ModelParameters,
) float64 {
	avgTOIMinutes := calculateAvgTOIMinutes(playerStats.Last5Games)

	// Base prediction starts from TOI rather than shot averages
	// Formula: adjust shot expectations based primarily on TOI
	// Players with 20+ minutes typically get 3-4 shots per game (first line/top pair)
	// Players with 15-20 minutes typically get 2-3 shots per game (second line/pair)
	// Players with 10-15 minutes typically get 1-2 shots per game (third line/pair)
	// Players with <10 minutes typically get 0-1 shots per game (fourth line/bottom pair)

	baseTOIPrediction := 0.0
	if avgTOIMinutes >= 20.0 {
		baseTOIPrediction = 3.5
	} else if avgTOIMinutes >= 15.0 {
		baseTOIPrediction = 2.5
	} else if avgTOIMinutes >= 10.0 {
		baseTOIPrediction = 1.5
	} else {
		baseTOIPrediction = 0.7
	}

	// Position adjustment is more significant in this model
	if playerStats.Position == "D" {
		if avgTOIMinutes >= 20.0 {
			// Top pair defensemen still get shots
			baseTOIPrediction *= params.DefensePositionFactor
		} else {
			// Lower pair defensemen get fewer shots
			baseTOIPrediction *= (params.DefensePositionFactor - 0.1)
		}
	}

	// Blend TOI prediction with actual shot history, but heavily favor TOI
	toiWeight := 0.7         // 70% TOI-based prediction
	shotHistoryWeight := 0.3 // 30% shot history

	blendedBasePrediction := (baseTOIPrediction * toiWeight) +
		((avgShotsLast5*params.RecentPerformanceWeight +
			seasonShotsPerGame*params.SeasonPerformanceWeight) * shotHistoryWeight)

	// Apply team factors with reduced weight
	currentTeam := getTeamStatsById(playerStats.CurrentTeamId, teamStats)
	opposingTeam := getTeamStatsById(playerStats.OpposingTeamId, teamStats)

	if currentTeam == nil || opposingTeam == nil {
		return math.Round(blendedBasePrediction*10) / 10
	}

	// Team pace factors with reduced impact
	leagueShotAverage := getLeagueShotAverage(teamStats)
	gamePaceFactor := math.Pow((currentTeam.ShotsForPerGame+opposingTeam.ShotsAgainstPerGame)/(2*leagueShotAverage), params.GamePaceExponent)

	// Team strength adjustments with reduced impact
	teamOffenseFactor := math.Pow(currentTeam.ShotsForPerGame/leagueShotAverage, params.TeamOffenseExponent)
	teamDefenseFactor := math.Pow(opposingTeam.ShotsAgainstPerGame/leagueShotAverage, params.TeamDefenseExponent)

	// Rest day impact with reduced factor
	restFactor := 1.0
	if days, exists := restDays[playerStats.CurrentTeamId]; exists {
		switch {
		case days == 0:
			restFactor = params.BackToBackFactor
		case days == 1:
			restFactor = params.OneRestDayFactor
		case days >= 4:
			restFactor = params.FourPlusRestDayFactor
		}
	}

	// Final prediction with team factors having less impact
	adjustedPrediction := blendedBasePrediction *
		math.Pow(gamePaceFactor, 0.7) *
		math.Pow(teamOffenseFactor, 0.7) *
		math.Pow(teamDefenseFactor, 0.7) *
		math.Pow(restFactor, 0.8)

	return math.Round(adjustedPrediction*10) / 10
}

// Matchup-focused calculation that emphasizes team matchups and contextual factors
func calculatePredictedShotsMatchupFocused(
	playerStats models.PlayerDetail,
	avgShotsLast5 float64,
	seasonShotsPerGame float64,
	teamStats []models.TeamStats,
	restDays map[int]int,
	params models.ModelParameters,
) float64 {
	currentTeam := getTeamStatsById(playerStats.CurrentTeamId, teamStats)
	opposingTeam := getTeamStatsById(playerStats.OpposingTeamId, teamStats)

	if currentTeam == nil || opposingTeam == nil {
		return calculatePredictedShotsStandard(playerStats, avgShotsLast5, seasonShotsPerGame, teamStats, restDays, params)
	}

	// Start with standard calculation weights but emphasize matchup
	basePrediction := avgShotsLast5*params.RecentPerformanceWeight + seasonShotsPerGame*params.SeasonPerformanceWeight

	// Enhanced team pace factor - emphasize matchup more
	leagueShotAverage := getLeagueShotAverage(teamStats)
	gamePaceFactor := math.Pow((currentTeam.ShotsForPerGame+opposingTeam.ShotsAgainstPerGame)/(2*leagueShotAverage), params.GamePaceExponent)

	// Enhanced team strength adjustments
	teamOffenseFactor := math.Pow(currentTeam.ShotsForPerGame/leagueShotAverage, params.TeamOffenseExponent)
	teamDefenseFactor := math.Pow(opposingTeam.ShotsAgainstPerGame/leagueShotAverage, params.TeamDefenseExponent)

	// Home ice advantage factor (if player's team is home)
	homeIceFactor := 1.0
	if playerStats.CurrentTeamId == playerStats.OpposingTeamId {
		// Opposing team ID will be different from current team ID for away games
		homeIceFactor = params.HomeIceAdvantageFactor
	}

	// Position-based adjustment with matchup context
	positionFactor := 1.0
	if playerStats.Position == "D" {
		// Check if opposing team shoots a lot
		if opposingTeam.ShotsForPerGame > leagueShotAverage {
			// Defensemen might block more shots against high-volume shooting teams
			positionFactor = params.DefensePositionFactor + 0.05
		} else {
			positionFactor = params.DefensePositionFactor
		}
	}

	// Calculate ice time factor as in standard model
	avgTOIMinutes := calculateAvgTOIMinutes(playerStats.Last5Games)
	icetimeFactor := math.Min(avgTOIMinutes/20.0, 1.0)

	// Enhanced rest day factors with more significance
	restFactor := 1.0
	if days, exists := restDays[playerStats.CurrentTeamId]; exists {
		switch {
		case days == 0:
			restFactor = params.BackToBackFactor
		case days == 1:
			restFactor = params.OneRestDayFactor
		case days >= 4:
			restFactor = params.FourPlusRestDayFactor
		}
	}

	// Team streak factor (placeholder - would require additional data)
	// This represents how a team's winning/losing streak might impact shot volume
	streakFactor := 1.0 + (params.StreakImpactFactor * 0) // Neutral for now

	// Combine all factors with enhanced matchup emphasis
	adjustedPrediction := basePrediction *
		math.Pow(gamePaceFactor, 1.2) * // Increased impact
		math.Pow(teamOffenseFactor, 1.1) * // Increased impact
		math.Pow(teamDefenseFactor, 1.1) * // Increased impact
		positionFactor *
		icetimeFactor *
		restFactor *
		homeIceFactor *
		streakFactor

	return math.Round(adjustedPrediction*10) / 10
}

// Strategy router that selects the appropriate calculation method based on model strategy
func calculatePredictedShots(
	playerStats models.PlayerDetail,
	shotsLast5 []int,
	seasonShotsPerGame float64,
	teamStats []models.TeamStats,
	restDays map[int]int,
	model models.ModelVersion,
) float64 {
	// Calculate average shots for standard methods
	var totalShots float64
	for _, shots := range shotsLast5 {
		totalShots += float64(shots)
	}
	avgShotsLast5 := totalShots / float64(len(shotsLast5))

	// Route to appropriate calculation method based on strategy
	switch model.CalculationStrategy {
	case models.StandardCalculation:
		return calculatePredictedShotsStandard(playerStats, avgShotsLast5, seasonShotsPerGame, teamStats, restDays, model.Parameters)
	case models.WeightedRecencyCalculation:
		return calculatePredictedShotsWeightedRecency(playerStats, shotsLast5, seasonShotsPerGame, teamStats, restDays, model.Parameters)
	case models.TOIDrivenCalculation:
		return calculatePredictedShotsTOIDriven(playerStats, avgShotsLast5, seasonShotsPerGame, teamStats, restDays, model.Parameters)
	case models.MatchupFocusedCalculation:
		return calculatePredictedShotsMatchupFocused(playerStats, avgShotsLast5, seasonShotsPerGame, teamStats, restDays, model.Parameters)
	default:
		// Fall back to standard calculation if strategy is unknown
		return calculatePredictedShotsStandard(playerStats, avgShotsLast5, seasonShotsPerGame, teamStats, restDays, model.Parameters)
	}
}

// CalculateShootingStatsWithModel calculates shooting stats using the specified model version
func CalculateShootingStatsWithModel(players []models.PlayerDetail, teamStats []models.TeamStats, restDays map[int]int, model models.ModelVersion) []models.PlayerStats {
	var stats []models.PlayerStats
	params := model.Parameters

	for _, player := range players {
		if len(player.Last5Games) < 5 {
			continue
		}

		// Only include players who have played in the last 7 days
		if player.Last5Games[0].GameDate < time.Now().AddDate(0, 0, -7).Format("2006-01-02") {
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

		predictedGameShots := calculatePredictedShots(player, shotsLast5, seasonShotsPerGame, teamStats, restDays, model)

		// Only include players meeting minimum shot threshold
		if predictedGameShots >= models.MIN_SHOTS {
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
				PredictedGameShots: predictedGameShots,
				Confidence:         calculateConfidence(predictedGameShots, avgTOI, shotTrend, player.Position, params),
				RestDays:           restDays[player.CurrentTeamId],
				Headshot:           player.Headshot,
				// Record which model was used
				ModelVersionID: model.ID,
			})
		}
	}

	// Sort by confidence
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Confidence > stats[j].Confidence
	})

	return stats
}

// CalculateShootingStats is the original function, now calling the versioned one with the active model
func CalculateShootingStats(players []models.PlayerDetail, teamStats []models.TeamStats, restDays map[int]int) []models.PlayerStats {
	model, err := GetModelVersion(GetActiveModelVersion())
	if err != nil {
		// Fallback to default model parameters if no model is found
		defaultModel := models.ModelVersion{
			ID:                  0,
			CalculationStrategy: models.StandardCalculation,
			Parameters: models.ModelParameters{
				RecentPerformanceWeight: 0.7,
				SeasonPerformanceWeight: 0.3,
				GamePaceExponent:        1.0,
				TeamOffenseExponent:     0.8,
				TeamDefenseExponent:     0.6,
				DefensePositionFactor:   0.75,
				BackToBackFactor:        0.9,
				OneRestDayFactor:        0.95,
				FourPlusRestDayFactor:   1.1,
				ShotScoreMultiplier:     4.0,
				TOIBaseMultiplier:       3.0,
				TOIBonusThreshold:       18.0,
				TOIBonusMultiplier:      1.0,
				TrendUpwardScore:        1.5,
				TrendImprovementScore:   0.75,
				DefenseConfidenceFactor: 0.9,
			},
		}
		return CalculateShootingStatsWithModel(players, teamStats, restDays, defaultModel)
	}

	return CalculateShootingStatsWithModel(players, teamStats, restDays, *model)
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

// GetGamesForModelPrediction gets upcoming games with player predictions using a specific model
func GetGamesForModelPrediction(date string, model models.ModelVersion) ([]models.GameWithPlayers, error) {
	games, err := repository.GetUpcomingGames(date)
	if err != nil {
		return nil, err
	}

	if len(games) == 0 {
		return nil, fmt.Errorf("no NHL games found for date")
	}

	teamStats, err := repository.GetAllTeamStats(games[0].Season)
	if err != nil {
		return nil, err
	}

	var restDays map[int]int
	restDays, err = repository.GetTeamsRest(date, games)
	if err != nil {
		return nil, err
	}

	var allPlayers []models.PlayerStats
	for _, game := range games {
		players, err := repository.GetPlayerStats(game.GameID, []models.Team{game.AwayTeam, game.HomeTeam})
		if err != nil {
			fmt.Println("Error fetching player stats:", err)
			continue
		}
		playerStats := CalculateShootingStatsWithModel(players, teamStats, restDays, model)
		allPlayers = append(allPlayers, playerStats...)
	}

	var gamesWithPlayers []models.GameWithPlayers

	for _, game := range games {
		filteredPlayers := helpers.Filter(allPlayers, func(player models.PlayerStats) bool {
			return player.TeamId == game.AwayTeam.Id || player.TeamId == game.HomeTeam.Id
		})

		mappedPlayers := helpers.Map(filteredPlayers, func(player models.PlayerStats) models.PlayerStats {
			// Don't load past prediction data since this is just a simulation
			return player
		})

		gamesWithPlayers = append(gamesWithPlayers, models.GameWithPlayers{
			Game:    game,
			Players: mappedPlayers,
		})
	}

	return gamesWithPlayers, nil
}

func GetPlayerShotStats(date string) ([]models.GameWithPlayers, error) {
	// Ensure models are initialized
	if !initialized {
		InitializeModels()
	}

	games, err := repository.GetUpcomingGames(date)
	if err != nil {
		return nil, err
	}

	if len(games) == 0 {
		return nil, fmt.Errorf("no NHL games found for date")
	}

	teamStats, err := repository.GetAllTeamStats(games[0].Season)
	if err != nil {
		return nil, err
	}

	var restDays map[int]int
	restDays, err = repository.GetTeamsRest(date, games)
	if err != nil {
		return nil, err
	}

	var allPlayers []models.PlayerStats
	for _, game := range games {
		players, err := repository.GetPlayerStats(game.GameID, []models.Team{game.AwayTeam, game.HomeTeam})
		if err != nil {
			fmt.Println("Error fetching player stats:", err)
			continue
		}
		playerStats := CalculateShootingStats(players, teamStats, restDays)
		allPlayers = append(allPlayers, playerStats...)
	}

	var gamesWithPlayers []models.GameWithPlayers

	for _, game := range games {
		filteredPlayers := helpers.Filter(allPlayers, func(player models.PlayerStats) bool {
			return player.TeamId == game.AwayTeam.Id || player.TeamId == game.HomeTeam.Id
		})

		mappedPlayers := helpers.Map(filteredPlayers, func(player models.PlayerStats) models.PlayerStats {
			player.PastPredictionAccuracy, err = dbRepository.GetPlayerPastPredictionAccuracy(player.PlayerId)
			if err != nil {
				fmt.Println("Error fetching player past prediction accuracy:", err)
			}

			predictionRecord, err := dbRepository.GetPlayerPredictionRecord(player.PlayerId, &game.GameID)
			if err != nil {
				fmt.Println("Error fetching player prediction record:", err)
			}
			predictionRecord.ModelVersionID = player.ModelVersionID
			player.PredictionRecord = predictionRecord
			return player
		})

		gamesWithPlayers = append(gamesWithPlayers, models.GameWithPlayers{
			Game:    game,
			Players: mappedPlayers,
		})
	}

	return gamesWithPlayers, nil
}
