package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"

	"api.alexmontague.ca/helpers"
	dbRepository "api.alexmontague.ca/internal/database/repository"
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
		date = helpers.GetCurrentESTDate()
	}

	fmt.Println("[nhl/shots] Fetching upcoming games for date:", date)

	gamesWithPlayers, err := service.GetPlayerShotStats(date)
	if err != nil {
		json.NewEncoder(w).Encode(helpers.Response{
			Error:   true,
			Code:    500,
			Message: fmt.Sprintf("Error fetching player shot stats: %s", err),
		})
		return
	}
	fmt.Printf("Total API requests made: %d\n", atomic.LoadUint64(&repository.RequestCount))

	totalAccuracy, err := dbRepository.GetTotalAccuracy()
	if err != nil {
		json.NewEncoder(w).Encode(helpers.Response{
			Error:   true,
			Code:    500,
			Message: fmt.Sprintf("Error fetching total accuracy: %s", err),
		})
		return
	}

	json.NewEncoder(w).Encode(struct {
		GamesWithPlayers []models.GameWithPlayers `json:"gamesWithPlayers"`
		TotalAccuracy    float64                  `json:"totalAccuracy"`
	}{
		GamesWithPlayers: gamesWithPlayers,
		TotalAccuracy:    totalAccuracy,
	})
}

// Route : '/nhl/shots/records?date=2025-02-23
// Type  : 'GET'
func GetPlayerShotRecords(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	date := r.URL.Query().Get("date")
	if date == "" {
		date = helpers.GetCurrentESTDate()
	}

	fmt.Println("[nhl/shots/records] Fetching shot records for date:", date)

	predictionRecords, err := dbRepository.GetGamePredictionsForDate(date)
	if err != nil {
		json.NewEncoder(w).Encode(helpers.Response{
			Error:   true,
			Code:    500,
			Message: fmt.Sprintf("Error fetching player shot stats: %s", err),
		})
		return
	}

	json.NewEncoder(w).Encode(predictionRecords)
}

// SeedAndValidatePredictions saves player predictions and updates with actual shots
// Route : '/nhl/shots/seed'
// Type  : 'GET'
func SeedAndValidatePredictions(w http.ResponseWriter, r *http.Request) {
	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Check for a specific date
	query := r.URL.Query()
	dateParam := query.Get("date")
	if dateParam == "" {
		dateParam = helpers.GetCurrentESTDate()
	}

	// Run predictions for all models for this date
	err := service.RunAndStoreAllModelPredictions(dateParam)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":    err.Error(),
			"gameDate": dateParam,
		})
		return
	}

	// Return success response
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"message":  "Predictions generated and stored for all models",
		"gameDate": dateParam,
	})
}
