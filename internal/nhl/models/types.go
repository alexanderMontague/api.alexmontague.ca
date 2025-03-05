package models

const NHL_API_BASE = "https://api-web.nhle.com/v1"
const NHL_STATS_API_BASE = "https://api.nhle.com/stats/rest/en"
const MIN_SHOTS = 2

type Team struct {
	Abbrev string `json:"abbrev"`
	Logo   string `json:"logo"`
	Id     int    `json:"id"`
}

type Game struct {
	GameID       int    `json:"gameId"`
	Title        string `json:"title"`
	AwayTeam     Team   `json:"awayTeam"`
	HomeTeam     Team   `json:"homeTeam"`
	Season       string `json:"season"`
	StartTimeUTC string `json:"startTimeUTC"`
	EstDate      string `json:"estDate"`
}

type PlayerStats struct {
	PlayerId           int     `json:"playerId"`
	Name               string  `json:"name"`
	Position           string  `json:"position"`
	TeamAbbrev         string  `json:"teamAbbrev"`
	TeamId             int     `json:"teamId"`
	ShotsLast5         []int   `json:"shotsLast5"`
	AvgShotsLast5      float64 `json:"avgShotsLast5"`
	ShotTrend          []int   `json:"shotTrend"`
	AvgTOI             float64 `json:"avgTOI"`
	SeasonShotsPerGame float64 `json:"seasonShotsPerGame"`
	PredictedGameShots float64 `json:"predictedGameShots"`
	Confidence         float64 `json:"confidence"`
	RestDays           int     `json:"restDays"`
	Headshot           string  `json:"headshot"`
}

type TeamStats struct {
	FaceoffWinPct       float64 `json:"faceoffWinPct"`
	GamesPlayed         int     `json:"gamesPlayed"`
	GoalsAgainst        int     `json:"goalsAgainst"`
	GoalsAgainstPerGame float64 `json:"goalsAgainstPerGame"`
	GoalsFor            int     `json:"goalsFor"`
	GoalsForPerGame     float64 `json:"goalsForPerGame"`
	Losses              int     `json:"losses"`
	OTLosses            int     `json:"otLosses"`
	PenaltyKillNetPct   float64 `json:"penaltyKillNetPct"`
	PenaltyKillPct      float64 `json:"penaltyKillPct"`
	PointPct            float64 `json:"pointPct"`
	Points              int     `json:"points"`
	PowerPlayNetPct     float64 `json:"powerPlayNetPct"`
	PowerPlayPct        float64 `json:"powerPlayPct"`
	RegulationAndOtWins int     `json:"regulationAndOtWins"`
	SeasonId            int     `json:"seasonId"`
	ShotsAgainstPerGame float64 `json:"shotsAgainstPerGame"`
	ShotsForPerGame     float64 `json:"shotsForPerGame"`
	TeamFullName        string  `json:"teamFullName"`
	TeamId              int     `json:"teamId"`
	Wins                int     `json:"wins"`
	WinsInRegulation    int     `json:"winsInRegulation"`
	WinsInShootout      int     `json:"winsInShootout"`
}

type ShotResponse struct {
	Date  string `json:"date"`
	Games []struct {
		Game
		Players []PlayerStats `json:"players"`
	} `json:"games"`
}

type NameField struct {
	Default string `json:"default"`
}

type Player struct {
	Id        int       `json:"id"`
	FirstName NameField `json:"firstName"`
	LastName  NameField `json:"lastName"`
	Position  string    `json:"positionCode"`
}

type RosterResponse struct {
	Forwards   []Player `json:"forwards"`
	Defensemen []Player `json:"defensemen"`
}

type Last5Game struct {
	Shots int    `json:"shots"`
	TOI   string `json:"toi"`
}

type PlayerDetail struct {
	PlayerId  int `json:"playerId"`
	FirstName struct {
		Default string `json:"default"`
	} `json:"firstName"`
	LastName struct {
		Default string `json:"default"`
	} `json:"lastName"`
	Position           string `json:"position"`
	CurrentTeamId      int    `json:"currentTeamId"`
	CurrentTeamAbbrev  string `json:"currentTeamAbbrev"`
	OpposingTeamId     int    `json:"opposingTeamId"`
	OpposingTeamAbbrev string `json:"opposingTeamAbbrev"`
	FeaturedStats      struct {
		RegularSeason struct {
			SubSeason struct {
				Shots       int `json:"shots"`
				GamesPlayed int `json:"gamesPlayed"`
			} `json:"subSeason"`
		} `json:"regularSeason"`
	} `json:"featuredStats"`
	Last5Games []Last5Game `json:"last5Games"`
	Headshot   string      `json:"headshot"`
}

type GamesWithPlayers struct {
	Game
	Players []PlayerStats `json:"players"`
}
