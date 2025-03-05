package cron

import (
	"fmt"
	"log"
	"math"
	"time"

	"api.alexmontague.ca/internal/database/repository"
	"api.alexmontague.ca/internal/nhl/models"
	nhlRepo "api.alexmontague.ca/internal/nhl/repository"
	"api.alexmontague.ca/internal/nhl/service"
	"github.com/robfig/cron/v3"
)

var scheduler *cron.Cron

// StartScheduler initializes and starts the cron scheduler
func StartScheduler() {
	scheduler = cron.New(cron.WithSeconds())

	// Fetch daily games and make predictions (run at midnight ET)
	scheduler.AddFunc("0 0 5 * * *", fetchDailyGamesWithRetry)

	// Validate completed games (run every 6 hours)
	scheduler.AddFunc("0 0 */6 * * *", validateCompletedGamesWithRetry)

	scheduler.Start()
}

// StopScheduler stops the scheduler
func StopScheduler() {
	if scheduler != nil {
		scheduler.Stop()
	}
}

// fetchDailyGamesWithRetry gets games with exponential backoff retry
func fetchDailyGamesWithRetry() {
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		if err := fetchDailyGames(); err != nil {
			backoff := time.Duration(math.Pow(2, float64(i))) * time.Minute
			log.Printf("Failed to fetch daily games (attempt %d/%d): %v. Retrying in %v...",
				i+1, maxRetries, err, backoff)
			time.Sleep(backoff)
			continue
		}
		return // Success
	}
	log.Printf("Failed to fetch daily games after %d attempts", maxRetries)
}

// fetchDailyGames gets today's games and stores predictions
func fetchDailyGames() error {
	today := time.Now().Format("2006-01-02")
	games, err := nhlRepo.GetUpcomingGames(today)
	if err != nil {
		return fmt.Errorf("failed to get upcoming games: %w", err)
	}

	// Store games in database
	for _, game := range games {
		if err := repository.InsertGame(game); err != nil {
			return fmt.Errorf("failed to insert game: %w", err)
		}
	}

	// Get team stats for the current season
	season := ""
	if len(games) > 0 {
		season = games[0].Season
	} else {
		// If no games today, use current year
		season = time.Now().Format("20062007")
	}

	teamStats, err := nhlRepo.GetAllTeamStats(season)
	if err != nil {
		return fmt.Errorf("failed to get team stats: %w", err)
	}

	// Get rest days for teams
	restDays, err := nhlRepo.GetTeamsRest(today, games)
	if err != nil {
		log.Printf("Warning: Failed to get rest days: %v", err)
		restDays = make(map[int]int) // Use empty map instead of failing
	}

	// Process each game
	for _, game := range games {
		// Get both teams for the game
		teamInfo := []models.Team{game.AwayTeam, game.HomeTeam}

		// Get player stats
		players, err := service.GetPlayerStats(game.GameID, teamInfo)
		if err != nil {
			return fmt.Errorf("failed to get player stats for game %d: %w", game.GameID, err)
		}

		// Calculate shot predictions
		predictions := service.CalculateShootingStats(players, teamStats, restDays)

		// Store predictions
		if err := repository.StorePredictions(game.GameID, predictions); err != nil {
			return fmt.Errorf("failed to store predictions for game %d: %w", game.GameID, err)
		}

		// Mark game as processed
		if err := repository.MarkGameProcessed(game.GameID); err != nil {
			return fmt.Errorf("failed to mark game as processed: %w", err)
		}
	}

	return nil
}

// validateCompletedGamesWithRetry validates games with exponential backoff retry
func validateCompletedGamesWithRetry() {
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		if err := validateCompletedGames(); err != nil {
			backoff := time.Duration(math.Pow(2, float64(i))) * time.Minute
			log.Printf("Failed to validate games (attempt %d/%d): %v. Retrying in %v...",
				i+1, maxRetries, err, backoff)
			time.Sleep(backoff)
			continue
		}
		return // Success
	}
	log.Printf("Failed to validate games after %d attempts", maxRetries)
}

// validateCompletedGames validates shot predictions against actual results
func validateCompletedGames() error {
	// Get completed games that need validation
	gameIDs, err := repository.GetCompletedUnvalidatedGames()
	if err != nil {
		return fmt.Errorf("failed to get unvalidated games: %w", err)
	}

	for _, gameID := range gameIDs {
		// Implement NHL API call to fetch actual shot results
		// This is hypothetical and would need to be implemented based on NHL API
		actualShots, err := fetchActualShots(gameID)
		if err != nil {
			return fmt.Errorf("failed to fetch actual shots for game %d: %w", gameID, err)
		}

		// Store actual results
		for playerID, shots := range actualShots {
			if err := repository.StoreActualShots(gameID, playerID, shots); err != nil {
				return fmt.Errorf("failed to store shot results: %w", err)
			}
		}
	}

	return nil
}

// fetchActualShots gets the actual shots for all players in a game
// This would need to be implemented using the NHL API
func fetchActualShots(gameID int) (map[int]int, error) {
	// This is a placeholder that should be replaced with actual NHL API implementation
	// Example structure: map[playerID]shotCount
	return map[int]int{}, fmt.Errorf("not implemented yet")
}
