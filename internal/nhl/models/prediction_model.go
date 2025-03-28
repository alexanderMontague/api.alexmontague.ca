package models

// CalculationStrategy defines the calculation approach used by a model
type CalculationStrategy string

const (
	// StandardCalculation uses the original shot prediction formula
	StandardCalculation CalculationStrategy = "standard"

	// WeightedRecencyCalculation places higher emphasis on recent performance trends
	WeightedRecencyCalculation CalculationStrategy = "weighted_recency"

	// TOIDrivenCalculation uses ice time as the primary predictor
	TOIDrivenCalculation CalculationStrategy = "toi_driven"

	// MatchupFocusedCalculation emphasizes team matchup statistics
	MatchupFocusedCalculation CalculationStrategy = "matchup_focused"
)

// ModelVersion represents a specific version of the shot prediction model
type ModelVersion struct {
	ID                  int
	Name                string
	Description         string
	CalculationStrategy CalculationStrategy
	Parameters          ModelParameters
	Active              bool
	CreatedAt           string
}

// ModelParameters contains all configurable parameters for a prediction model
type ModelParameters struct {
	// Weight factors
	RecentPerformanceWeight float64
	SeasonPerformanceWeight float64

	// Team and game factors
	GamePaceExponent    float64
	TeamOffenseExponent float64
	TeamDefenseExponent float64

	// Position adjustment
	DefensePositionFactor float64

	// Rest day factors
	BackToBackFactor      float64
	OneRestDayFactor      float64
	FourPlusRestDayFactor float64

	// Confidence calculation parameters
	ShotScoreMultiplier     float64
	TOIBaseMultiplier       float64
	TOIBonusThreshold       float64
	TOIBonusMultiplier      float64
	TrendUpwardScore        float64
	TrendImprovementScore   float64
	DefenseConfidenceFactor float64

	// Advanced parameters
	LastGameWeight              float64
	SecondLastGameWeight        float64
	ThirdLastGameWeight         float64
	FourthLastGameWeight        float64
	FifthLastGameWeight         float64
	OpposingGoalieQualityFactor float64
	HomeIceAdvantageFactor      float64
	StreakImpactFactor          float64
}

// GetDefaultModels returns the predefined model versions
func GetDefaultModels() []ModelVersion {
	return []ModelVersion{
		{
			ID:                  1,
			Name:                "Original Model",
			Description:         "The initial shot prediction model",
			CalculationStrategy: StandardCalculation,
			Active:              true,
			CreatedAt:           "2025-03-11",
			Parameters: ModelParameters{
				// Original weights
				RecentPerformanceWeight: 0.7,
				SeasonPerformanceWeight: 0.3,

				// Team and game factors
				GamePaceExponent:    1.0,
				TeamOffenseExponent: 0.8,
				TeamDefenseExponent: 0.6,

				// Position adjustment
				DefensePositionFactor: 0.75,

				// Rest day factors
				BackToBackFactor:      0.9,
				OneRestDayFactor:      0.95,
				FourPlusRestDayFactor: 1.1,

				// Confidence calculation parameters
				ShotScoreMultiplier:     4.0,
				TOIBaseMultiplier:       3.0,
				TOIBonusThreshold:       18.0,
				TOIBonusMultiplier:      1.0,
				TrendUpwardScore:        1.5,
				TrendImprovementScore:   0.75,
				DefenseConfidenceFactor: 0.9,

				// Advanced parameters not used in standard calculation
				LastGameWeight:              0.0,
				SecondLastGameWeight:        0.0,
				ThirdLastGameWeight:         0.0,
				FourthLastGameWeight:        0.0,
				FifthLastGameWeight:         0.0,
				OpposingGoalieQualityFactor: 0.0,
				HomeIceAdvantageFactor:      0.0,
				StreakImpactFactor:          0.0,
			},
		},
		{
			ID:                  2,
			Name:                "Enhanced Recent Performance",
			Description:         "Increased weight on recent performance and rest factors",
			CalculationStrategy: StandardCalculation,
			Active:              false,
			CreatedAt:           "2025-03-27",
			Parameters: ModelParameters{
				// Modified weights
				RecentPerformanceWeight: 0.8,
				SeasonPerformanceWeight: 0.2,

				// Team and game factors
				GamePaceExponent:    1.0,
				TeamOffenseExponent: 0.9,
				TeamDefenseExponent: 0.7,

				// Position adjustment
				DefensePositionFactor: 0.8,

				// Rest day factors
				BackToBackFactor:      0.85,
				OneRestDayFactor:      0.95,
				FourPlusRestDayFactor: 1.15,

				// Confidence calculation parameters
				ShotScoreMultiplier:     4.0,
				TOIBaseMultiplier:       3.2,
				TOIBonusThreshold:       18.0,
				TOIBonusMultiplier:      1.1,
				TrendUpwardScore:        1.6,
				TrendImprovementScore:   0.8,
				DefenseConfidenceFactor: 0.9,

				// Advanced parameters not used in standard calculation
				LastGameWeight:              0.0,
				SecondLastGameWeight:        0.0,
				ThirdLastGameWeight:         0.0,
				FourthLastGameWeight:        0.0,
				FifthLastGameWeight:         0.0,
				OpposingGoalieQualityFactor: 0.0,
				HomeIceAdvantageFactor:      0.0,
				StreakImpactFactor:          0.0,
			},
		},
		{
			ID:                  3,
			Name:                "Weighted Recency Model",
			Description:         "Weighs individual games based on recency",
			CalculationStrategy: WeightedRecencyCalculation,
			Active:              false,
			CreatedAt:           "2025-03-27",
			Parameters: ModelParameters{
				// Base weights
				RecentPerformanceWeight: 0.85,
				SeasonPerformanceWeight: 0.15,

				// Team and game factors
				GamePaceExponent:    1.1,
				TeamOffenseExponent: 0.8,
				TeamDefenseExponent: 0.7,

				// Position adjustment
				DefensePositionFactor: 0.75,

				// Rest day factors
				BackToBackFactor:      0.9,
				OneRestDayFactor:      0.95,
				FourPlusRestDayFactor: 1.08,

				// Confidence calculation parameters
				ShotScoreMultiplier:     4.0,
				TOIBaseMultiplier:       3.0,
				TOIBonusThreshold:       18.0,
				TOIBonusMultiplier:      1.0,
				TrendUpwardScore:        1.5,
				TrendImprovementScore:   0.75,
				DefenseConfidenceFactor: 0.9,

				// Advanced parameters used in weighted recency calculation
				LastGameWeight:              0.4,  // 40% weight to most recent game
				SecondLastGameWeight:        0.25, // 25% weight to second most recent game
				ThirdLastGameWeight:         0.15, // 15% weight to third most recent game
				FourthLastGameWeight:        0.1,  // 10% weight to fourth most recent game
				FifthLastGameWeight:         0.1,  // 10% weight to fifth most recent game
				OpposingGoalieQualityFactor: 0.0,
				HomeIceAdvantageFactor:      0.0,
				StreakImpactFactor:          0.0,
			},
		},
		{
			ID:                  4,
			Name:                "Matchup-Focused Model",
			Description:         "Emphasizes team matchups and contextual factors",
			CalculationStrategy: MatchupFocusedCalculation,
			Active:              false,
			CreatedAt:           "2025-03-27",
			Parameters: ModelParameters{
				// Base weights
				RecentPerformanceWeight: 0.6,
				SeasonPerformanceWeight: 0.4,

				// Team and game factors (enhanced for matchup focus)
				GamePaceExponent:    1.2,
				TeamOffenseExponent: 1.0,
				TeamDefenseExponent: 0.9,

				// Position adjustment
				DefensePositionFactor: 0.8,

				// Rest day factors (enhanced importance)
				BackToBackFactor:      0.85,
				OneRestDayFactor:      0.92,
				FourPlusRestDayFactor: 1.15,

				// Confidence calculation parameters
				ShotScoreMultiplier:     3.5,
				TOIBaseMultiplier:       3.0,
				TOIBonusThreshold:       17.5,
				TOIBonusMultiplier:      1.0,
				TrendUpwardScore:        1.2,
				TrendImprovementScore:   0.6,
				DefenseConfidenceFactor: 0.9,

				// Advanced parameters
				LastGameWeight:              0.0,
				SecondLastGameWeight:        0.0,
				ThirdLastGameWeight:         0.0,
				FourthLastGameWeight:        0.0,
				FifthLastGameWeight:         0.0,
				OpposingGoalieQualityFactor: 1.1,  // 10% adjustment based on goalie quality
				HomeIceAdvantageFactor:      1.05, // 5% boost for home games
				StreakImpactFactor:          0.05, // Small adjustment for team streaks
			},
		},
		{
			ID:                  5,
			Name:                "TOI-Driven Model",
			Description:         "Uses ice time as primary predictor with minimal adjustment factors",
			CalculationStrategy: TOIDrivenCalculation,
			Active:              false,
			CreatedAt:           "2025-03-27",
			Parameters: ModelParameters{
				// Base weights (less important in this model)
				RecentPerformanceWeight: 0.5,
				SeasonPerformanceWeight: 0.5,

				// Team and game factors (reduced importance)
				GamePaceExponent:    0.7,
				TeamOffenseExponent: 0.5,
				TeamDefenseExponent: 0.4,

				// Position adjustment
				DefensePositionFactor: 0.8,

				// Rest day factors
				BackToBackFactor:      0.92,
				OneRestDayFactor:      0.97,
				FourPlusRestDayFactor: 1.05,

				// Confidence calculation parameters (enhanced TOI influence)
				ShotScoreMultiplier:     3.0,
				TOIBaseMultiplier:       4.5,  // Much higher TOI impact
				TOIBonusThreshold:       16.0, // Lower threshold for TOI bonus
				TOIBonusMultiplier:      1.5,  // Higher bonus multiplier
				TrendUpwardScore:        1.0,
				TrendImprovementScore:   0.5,
				DefenseConfidenceFactor: 0.95,

				// Advanced parameters
				LastGameWeight:              0.0,
				SecondLastGameWeight:        0.0,
				ThirdLastGameWeight:         0.0,
				FourthLastGameWeight:        0.0,
				FifthLastGameWeight:         0.0,
				OpposingGoalieQualityFactor: 0.0,
				HomeIceAdvantageFactor:      0.0,
				StreakImpactFactor:          0.0,
			},
		},
	}
}

// Extend PlayerStats to include the model version used for prediction
type ModelPredictionRecord struct {
	ModelVersionID int
	Accuracy       float64
	HitTarget      bool
	ActualShots    int
	PredictedShots float64
}

// ModelAccuracyStats holds aggregated statistics about a model's performance
type ModelAccuracyStats struct {
	TotalPredictions      int
	SuccessfulPredictions int
	Accuracy              float64
	AvgError              float64
}
