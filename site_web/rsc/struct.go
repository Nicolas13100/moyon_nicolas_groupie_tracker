package API

// TwitchTokenResponse represents the structure of the response from the Twitch token endpoint
type TwitchTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

// Game represents the structure of the JSON response from IGDB
type Game struct {
	ID                    int   `json:"id"`
	AgeRatings            []int `json:"age_ratings"`
	Artworks              []int `json:"artworks"`
	Category              int   `json:"category"`
	Cover                 int   `json:"cover"`
	CoverLink             string
	CreatedAt             int   `json:"created_at"`
	ExternalGames         []int `json:"external_games"`
	FirstReleaseDate      int64 `json:"first_release_date"`
	FirstReleaseDateHuman string
	GameModes             []int `json:"game_modes"`
	Genres                []int `json:"genres"`
	GenresString          []string
	Name                  string `json:"name"`
	Platforms             []int  `json:"platforms"`
	ReleaseDates          []int  `json:"release_dates"`
	Screenshots           []int  `json:"screenshots"`
	ScreenshotsLink       string
	SimilarGames          []int  `json:"similar_games"`
	Slug                  string `json:"slug"`
	Summary               string `json:"summary"`
	Tags                  []int  `json:"tags"`
	UpdatedAt             int    `json:"updated_at"`
	URL                   string `json:"url"`
	VersionParent         int    `json:"version_parent"`
	VersionTitle          string `json:"version_title"`
	Checksum              string `json:"checksum"`
}

type TemplateData struct {
	RecommendedGames []Game
	LastGames        []Game
}
