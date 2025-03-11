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
		if err := fetchDailyPredictions(nil); err != nil {
			log.Printf("Error in daily prediction job: %v", err)
			// Retry after 30 minutes if failed
			time.AfterFunc(30*time.Minute, func() {
				log.Println("Retrying daily prediction job")
				if err := fetchDailyPredictions(nil); err != nil {
					log.Printf("Retry failed: %v", err)
				}
			})
		}
	})

	// Hourly job to validate completed games - runs every hour
	scheduler.AddFunc("0 0 * * * *", func() {
		log.Println("Running validation job")
		if err := validateCompletedGames(nil); err != nil {
			log.Printf("Error in validation job: %v", err)
			// Retry after 15 minutes if failed
			time.AfterFunc(15*time.Minute, func() {
				log.Println("Retrying validation job")
				if err := validateCompletedGames(nil); err != nil {
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

// fetchDailyPredictions fetches games for today and makes predictions
func fetchDailyPredictions(date *string) error {
	// Get today's date in EST
	est, _ := time.LoadLocation("America/New_York")
	var today string
	if date == nil {
		today = time.Now().In(est).Format("2006-01-02")
	} else {
		today = *date
	}

	// Get games with predictions
	gamesWithPlayers, err := service.GetPlayerShotStats(today)
	if err != nil {
		return fmt.Errorf("failed to get shot stats: %w", err)
	}

	// Store each game prediction
	for _, gameWithPlayers := range gamesWithPlayers {

		if err := repository.StoreGamePredictions(gameWithPlayers); err != nil {
			log.Printf("Error storing predictions for game %d: %v", gameWithPlayers.Game.GameID, err)
			continue
		}

		log.Printf("Stored predictions for game %d: %s", gameWithPlayers.Game.GameID, gameWithPlayers.Game.Title)
	}

	return nil
}

// validateCompletedGames validates predictions for completed games
func validateCompletedGames(date *string) error {
	est, _ := time.LoadLocation("America/New_York")
	var today string
	if date == nil {
		today = time.Now().In(est).Format("2006-01-02")
	} else {
		today = *date
	}
	predictionRecords, err := repository.GetGamePredictionsForDate(today)
	log.Printf("Found %d predictions for date %s", len(predictionRecords), today)
	if err != nil {
		return fmt.Errorf("failed to get pending games: %w", err)
	}

	// Get actual shots for all games
	gameIDs := make([]int, len(predictionRecords))
	for i, prediction := range predictionRecords {
		gameIDs[i] = prediction.GameID
	}
	actualShots, err := nhlRepo.FetchActualGameShots(gameIDs)
	if err != nil {
		log.Printf("Error fetching results for games %v: %v", gameIDs, err)
		return fmt.Errorf("failed to fetch actual shots: %w", err)
	}

	for _, prediction := range predictionRecords {
		if err := repository.StoreActualShots(prediction, actualShots[prediction.PlayerID]); err != nil {
			log.Printf("Error storing results for game %d: %v", prediction.GameID, err)
			continue
		}

		log.Printf("Validated game %d", prediction.GameID)
	}

	return nil
}
