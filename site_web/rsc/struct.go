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
	Logged           bool
}

type GameFull struct {
	ID                    int     `json:"id"`
	AgeRatings            []int   `json:"age_ratings"`
	AggregatedRating      float64 `json:"aggregated_rating"`
	AggregatedRatingCount int     `json:"aggregated_rating_count"`
	AlternativeNames      []int   `json:"alternative_names"`
	Artworks              []int   `json:"artworks"`
	Bundles               []int   `json:"bundles"`
	Category              int     `json:"category"`
	Cover                 int     `json:"cover"`
	CoverLink             string
	CreatedAt             int64 `json:"created_at"`
	DLCs                  []int `json:"dlcs"`
	Expansions            []int `json:"expansions"`
	ExternalGames         []int `json:"external_games"`
	FirstReleaseDate      int64 `json:"first_release_date"`
	FirstReleaseDateHuman string
	Follows               int   `json:"follows"`
	Franchises            []int `json:"franchises"`
	GameEngines           []int `json:"game_engines"`
	GameModes             []int `json:"game_modes"`
	Genres                []int `json:"genres"`
	GenresString          []string
	Hypes                 int     `json:"hypes"`
	InvolvedCompanies     []int   `json:"involved_companies"`
	Keywords              []int   `json:"keywords"`
	Name                  string  `json:"name"`
	Platforms             []int   `json:"platforms"`
	PlayerPerspectives    []int   `json:"player_perspectives"`
	Rating                float64 `json:"rating"`
	RatingCount           int     `json:"rating_count"`
	ReleaseDates          []int   `json:"release_dates"`
	Screenshots           []int   `json:"screenshots"`
	ScreenshotsLink       string
	SimilarGames          []int   `json:"similar_games"`
	Slug                  string  `json:"slug"`
	Storyline             string  `json:"storyline"`
	Summary               string  `json:"summary"`
	Tags                  []int   `json:"tags"`
	Themes                []int   `json:"themes"`
	TotalRating           float64 `json:"total_rating"`
	TotalRatingCount      int     `json:"total_rating_count"`
	UpdatedAt             int64   `json:"updated_at"`
	URL                   string  `json:"url"`
	Videos                []int   `json:"videos"`
	Websites              []int   `json:"websites"`
	Checksum              string  `json:"checksum"`
	LanguageSupports      []int   `json:"language_supports"`
	GameLocalizations     []int   `json:"game_localizations"`
	Collections           []int   `json:"collections"`
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// UserData structure for individual user data
type UserData struct {
	Fav []int `json:"fav"`
}

type CombinedData struct {
	Result interface{}
	Name   string
	Logged bool
	Fav    bool
}
