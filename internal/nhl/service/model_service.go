package service

import (
	"fmt"
	"sort"
	"sync"

	dbRepository "api.alexmontague.ca/internal/database/repository"
	"api.alexmontague.ca/internal/nhl/models"
	"api.alexmontague.ca/internal/nhl/repository"
)

const (
	// DEFAULT_MODEL_VERSION is the model version to use by default
	DEFAULT_MODEL_VERSION = 1
)

var (
	// activeModelVersion is the currently active model version
	activeModelVersion = DEFAULT_MODEL_VERSION

	// predictionModels contains all available models
	predictionModels []models.ModelVersion

	// mu prevents race conditions when accessing models
	mu sync.RWMutex

	// initialized flag to prevent multiple initializations
	initialized = false
)

// InitializeModels loads the default models on startup
// This is called from main.go and ensures models are only initialized once
func InitializeModels() {
	mu.Lock()
	defer mu.Unlock()

	if initialized {
		return
	}

	// Load default models
	predictionModels = models.GetDefaultModels()
	initialized = true
}

// GetActiveModelVersion returns the currently active model version ID
func GetActiveModelVersion() int {
	return activeModelVersion
}

// SetActiveModelVersion changes the active model version
func SetActiveModelVersion(versionID int) {
	mu.Lock()
	defer mu.Unlock()
	activeModelVersion = versionID
}

// GetModelVersion returns a specific model version by ID
// Falls back to default model if requested model is not found
func GetModelVersion(versionID int) (*models.ModelVersion, error) {
	// Initialize models if not already done
	if !initialized {
		InitializeModels()
	}

	mu.RLock()
	defer mu.RUnlock()

	// Try to find the requested model
	for _, model := range predictionModels {
		if model.ID == versionID {
			return &model, nil
		}
	}

	// If model not found, return the default model (one we know exists)
	for _, model := range predictionModels {
		if model.ID == DEFAULT_MODEL_VERSION {
			return &model, nil
		}
	}

	// Create a minimal default model if nothing else available
	return createDefaultModel(), nil
}

// GetAllModels returns all available model versions
func GetAllModels() []models.ModelVersion {
	// Initialize models if not already done
	if !initialized {
		InitializeModels()
	}

	mu.RLock()
	defer mu.RUnlock()

	// Return a copy to prevent modification of internal state
	modelsCopy := make([]models.ModelVersion, len(predictionModels))
	copy(modelsCopy, predictionModels)

	return modelsCopy
}

// CalculateWithAllModels runs predictions using all models for comparison
func CalculateWithAllModels(date string) (map[int][]models.GameWithPlayers, error) {
	// Initialize models if not already done
	if !initialized {
		InitializeModels()
	}

	mu.RLock()
	allModels := make([]models.ModelVersion, len(predictionModels))
	copy(allModels, predictionModels)
	mu.RUnlock()

	// Fetch game data only once for reuse
	games, teamStats, restDays, err := fetchGameDataForDate(date)
	if err != nil {
		return nil, err
	}

	if len(games) == 0 {
		return map[int][]models.GameWithPlayers{}, nil
	}

	// Calculate predictions with each model
	results := make(map[int][]models.GameWithPlayers)

	for _, model := range allModels {
		modelResults, err := ModelPredictionForGames(date, model, games, teamStats, restDays)
		if err != nil {
			fmt.Println("Failed to model prediction for model version:", model.ID, err)
			continue // Skip failed models
		}
		results[model.ID] = modelResults
	}

	return results, nil
}

// ModelPredictionForGames gets predictions for games with a specific model, reusing data if provided
func ModelPredictionForGames(
	date string,
	model models.ModelVersion,
	cachedGames []models.Game,
	cachedTeamStats []models.TeamStats,
	cachedRestDays map[int]int,
) ([]models.GameWithPlayers, error) {
	var games []models.Game
	var teamStats []models.TeamStats
	var restDays map[int]int
	var err error

	// Use cached data if provided, otherwise fetch from repository
	if len(cachedGames) > 0 {
		games = cachedGames
	} else {
		games, teamStats, restDays, err = fetchGameDataForDate(date)
		if err != nil {
			return nil, err
		}
	}

	if len(games) == 0 {
		return nil, nil
	}

	if cachedTeamStats != nil {
		teamStats = cachedTeamStats
	} else if teamStats == nil {
		teamStats, err = repository.GetAllTeamStats(games[0].Season)
		if err != nil {
			return nil, err
		}
	}

	if cachedRestDays != nil {
		restDays = cachedRestDays
	} else if restDays == nil {
		restDays, err = repository.GetTeamsRest(date, games)
		if err != nil {
			return nil, err
		}
	}

	// Calculate predictions with the specified model
	var allPlayers []models.PlayerStats
	for _, game := range games {
		players, err := repository.GetPlayerStats(game.GameID, []models.Team{game.AwayTeam, game.HomeTeam})
		if err != nil {
			continue
		}
		playerStats := CalculateShootingStatsWithModel(players, teamStats, restDays, model)
		allPlayers = append(allPlayers, playerStats...)
	}

	// Group players by game
	var gamesWithPlayers []models.GameWithPlayers
	for _, game := range games {
		// Filter players for this game
		var gamePlayers []models.PlayerStats
		for _, player := range allPlayers {
			if player.TeamId == game.AwayTeam.Id || player.TeamId == game.HomeTeam.Id {
				gamePlayers = append(gamePlayers, player)
			}
		}

		// Sort players by confidence (highest first)
		sortPlayersByConfidence(gamePlayers)

		gamesWithPlayers = append(gamesWithPlayers, models.GameWithPlayers{
			Game:    game,
			Players: gamePlayers,
		})
	}

	return gamesWithPlayers, nil
}

// Helper to sort players by confidence
func sortPlayersByConfidence(players []models.PlayerStats) {
	sort.Slice(players, func(i, j int) bool {
		return players[i].Confidence > players[j].Confidence
	})
}

// Helper to fetch game data for a specific date
func fetchGameDataForDate(date string) ([]models.Game, []models.TeamStats, map[int]int, error) {
	games, err := repository.GetUpcomingGames(date)
	if err != nil {
		return nil, nil, nil, err
	}

	if len(games) == 0 {
		return games, nil, nil, nil
	}

	teamStats, err := repository.GetAllTeamStats(games[0].Season)
	if err != nil {
		return games, nil, nil, err
	}

	restDays, err := repository.GetTeamsRest(date, games)
	if err != nil {
		return games, teamStats, nil, err
	}

	return games, teamStats, restDays, nil
}

// createDefaultModel provides a minimal default model when no others are available
func createDefaultModel() *models.ModelVersion {
	return &models.ModelVersion{
		ID:                  DEFAULT_MODEL_VERSION,
		Name:                "Default Model",
		Description:         "Basic shot prediction model",
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
}

// RunAndStoreAllModelPredictions runs all models and stores their predictions in the database
func RunAndStoreAllModelPredictions(date string) error {
	// Initialize models if not already done
	if !initialized {
		InitializeModels()
	}

	// Get predictions from all models
	modelPredictions, err := CalculateWithAllModels(date)
	if err != nil {
		return fmt.Errorf("failed to calculate predictions: %w", err)
	}

	// Store all model predictions
	err = dbRepository.StoreModelPredictions(date, modelPredictions)
	if err != nil {
		return fmt.Errorf("failed to store model predictions: %w", err)
	}

	// Also store the active model's prediction in the original table for backward compatibility
	if activeModelPredictions, ok := modelPredictions[activeModelVersion]; ok && len(activeModelPredictions) > 0 {
		for _, game := range activeModelPredictions {
			err = dbRepository.StoreGamePredictions(game)
			if err != nil {
				return fmt.Errorf("failed to store active model predictions: %w", err)
			}
		}
	}

	return nil
}
