package cron

import (
	"fmt"
	"log"
	"time"

	"api.alexmontague.ca/internal/database/repository"
	nhlRepo "api.alexmontague.ca/internal/nhl/repository"
	"api.alexmontague.ca/internal/nhl/service"
	"github.com/robfig/cron/v3"
)

var scheduler *cron.Cron

// StartScheduler initializes and starts the cron scheduler
func StartScheduler() {
	scheduler = cron.New(cron.WithSeconds())

	// Daily job to fetch games and make predictions - runs at 4:00 AM EST
	scheduler.AddFunc("0 0 4 * * *", func() {
		log.Println("Running daily prediction job")
		if err := FetchDailyPredictions(nil); err != nil {
			log.Printf("Error in daily prediction job: %v", err)
			// Retry after 30 minutes if failed
			time.AfterFunc(30*time.Minute, func() {
				log.Println("Retrying daily prediction job")
				if err := FetchDailyPredictions(nil); err != nil {
					log.Printf("Retry failed: %v", err)
				}
			})
		}
	})

	// Hourly job to validate completed games - runs every hour
	scheduler.AddFunc("0 0 * * * *", func() {
		log.Println("Running validation job")
		if err := ValidateCompletedGames(nil); err != nil {
			log.Printf("Error in validation job: %v", err)
			// Retry after 15 minutes if failed
			time.AfterFunc(15*time.Minute, func() {
				log.Println("Retrying validation job")
				if err := ValidateCompletedGames(nil); err != nil {
					log.Printf("Retry failed: %v", err)
				}
			})
		}
	})

	scheduler.Start()
	log.Println("Scheduler started")
}

// StopScheduler stops the cron scheduler
func StopScheduler() {
	if scheduler != nil {
		scheduler.Stop()
		log.Println("Scheduler stopped")
	}
}

// FetchDailyPredictions fetches games for today and makes predictions using all models
func FetchDailyPredictions(date *string) error {
	// Get today's date in EST
	est, _ := time.LoadLocation("America/New_York")
	var today string
	if date == nil {
		today = time.Now().In(est).Format("2006-01-02")
	} else {
		today = *date
	}

	// Run all models and store predictions in both tables
	err := service.RunAndStoreAllModelPredictions(today)
	if err != nil {
		return fmt.Errorf("failed to run and store model predictions: %w", err)
	}

	log.Printf("Successfully stored predictions for all models for date %s", today)
	return nil
}

// ValidateCompletedGames validates predictions for completed games in both tables
func ValidateCompletedGames(date *string) error {
	est, _ := time.LoadLocation("America/New_York")
	var today string
	if date == nil {
		today = time.Now().In(est).Format("2006-01-02")
	} else {
		today = *date
	}

	// Get predictions from original table
	predictionRecords, err := repository.GetGamePredictionsForDate(today)
	log.Printf("Found %d predictions in original table for date %s", len(predictionRecords), today)
	if err != nil {
		return fmt.Errorf("failed to get pending games: %w", err)
	}

	// Get actual shots for games
	gameIDs := make([]int, 0)
	gameIDMap := make(map[int]bool)

	for _, prediction := range predictionRecords {
		if !gameIDMap[prediction.GameID] {
			gameIDs = append(gameIDs, prediction.GameID)
			gameIDMap[prediction.GameID] = true
		}
	}

	if len(gameIDs) == 0 {
		log.Printf("No games to validate for date %s", today)
		return nil
	}

	// Fetch actual shots
	actualShots, err := nhlRepo.FetchActualGameShots(gameIDs)
	if err != nil {
		log.Printf("Error fetching results for games %v: %v", gameIDs, err)
		return fmt.Errorf("failed to fetch actual shots: %w", err)
	}

	// Update original table
	for _, prediction := range predictionRecords {
		// Skip if no actual shots data available for this player
		if _, exists := actualShots[prediction.PlayerID]; !exists {
			continue
		}

		if err := repository.StoreActualShots(prediction, actualShots[prediction.PlayerID]); err != nil {
			log.Printf("Error storing results for game %d, player %d in original table: %v",
				prediction.GameID, prediction.PlayerID, err)
			continue
		}

		// Also update in model_predictions table
		if err := repository.UpdateModelPredictionsWithActual(prediction.GameID, prediction.PlayerID, actualShots[prediction.PlayerID]); err != nil {
			log.Printf("Error storing results for game %d, player %d in model_predictions table: %v",
				prediction.GameID, prediction.PlayerID, err)
			continue
		}
	}

	log.Printf("Validated predictions for %d games", len(gameIDs))
	return nil
}
