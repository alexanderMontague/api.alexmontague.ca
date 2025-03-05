package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"api.alexmontague.ca/helpers"
	"api.alexmontague.ca/internal/nhl/repository"
	"api.alexmontague.ca/internal/nhl/service"
)

// Route : '/nhl/shots?date=2025-02-23
// Type  : 'GET'
func GetPlayerShotStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	date := r.URL.Query().Get("date")
	if date == "" {
		now := time.Now()
		loc, _ := time.LoadLocation("America/New_York")
		date = now.In(loc).Format("2006-01-02")
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

	json.NewEncoder(w).Encode(gamesWithPlayers)
}
