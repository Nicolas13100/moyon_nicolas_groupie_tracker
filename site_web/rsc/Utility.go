package API

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"
)

var (
	// GenreMap is a mapping of genre IDs to their names
	GenreMap     = make(map[int]string)
	genreMapOnce sync.Once
)

func renderTemplate(w http.ResponseWriter, tmplName string, data interface{}) {
	// Taken from hangman
	tmpl, err := template.New(tmplName).Funcs(template.FuncMap{"join": join}).ParseFiles("site_web/Template/" + tmplName + ".html")
	if err != nil {
		fmt.Println("Error parsing template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, tmplName, data)
	if err != nil {
		fmt.Println("Error executing template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func join(s []string, sep string) string {
	// same
	return strings.Join(s, sep)
}

func fetchRecommendedGames() ([]Game, error) {
	apiKey := "3pfgrdttfa66z525wc6d40uzjv9nq3" // the client ID
	apiURL := "https://api.igdb.com/v4/games"
	accessToken, err := GetTwitchOAuthToken()
	// keep in mind that max number of token at same time is 25 , oldest one is deleted after
	if err != nil {
		fmt.Println("error token", err)
		return nil, err
	}
	if accessToken == "" {
		fmt.Println("no token was given back")
	}
	// theme 42 MUST be out , it's erotica theme
	params := `
	fields *;
	where themes != 42 & rating > 65 & total_rating_count > 60;
	limit 15;
    sort rating desc;
	`

	// Make a POST request to the IGDB API with the parameters in the body
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(params))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Client-ID", apiKey)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON response
	var games []Game
	err = json.Unmarshal(body, &games)
	if err != nil {
		return nil, err
	}

	// Populate struct for each game
	for i := range games {
		games[i].GenresString = make([]string, len(games[i].Genres))
		for j, genreID := range games[i].Genres {
			games[i].GenresString[j] = GetGenreNameByID(genreID)
		}
	}
	// Populate struct for each game
	for y := range games {
		games[y].CoverLink = getCoverImageURL(games[y].Cover)
		games[y].FirstReleaseDateHuman = formatUnixTimestampToFrenchDate(games[y].FirstReleaseDate)
	}

	return games, nil
}

func GetTwitchOAuthToken() (string, error) {
	clientID := "3pfgrdttfa66z525wc6d40uzjv9nq3"
	clientSecret := "yzco2f5hb5473bbru0ltykc0lc5dv8"
	// Twitch OAuth token endpoint
	tokenURL := "https://id.twitch.tv/oauth2/token"

	// Define the request payload
	payload := map[string]string{
		"client_id":     clientID,
		"client_secret": clientSecret,
		"grant_type":    "client_credentials",
	}

	// Convert payload to JSON
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	// Make a POST request to the Twitch token endpoint
	resp, err := http.Post(tokenURL, "application/json", bytes.NewBuffer(payloadJSON))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Unmarshal the JSON response
	var tokenResponse TwitchTokenResponse
	err = json.Unmarshal(body, &tokenResponse)
	if err != nil {
		return "", err
	}

	// Return the access token
	return tokenResponse.AccessToken, nil
}

func initGenreMap() {
	apiURL := "https://api.igdb.com/v4/genres"
	apiKey := "3pfgrdttfa66z525wc6d40uzjv9nq3" // the client ID
	accessToken, err := GetTwitchOAuthToken()
	if err != nil {
		fmt.Println("error token", err)
	}
	if accessToken == "" {
		fmt.Println("no token was given back")
	}

	params := `
	fields id;
	fields name;
	limit 500;
	`

	// Make a POST request to the IGDB API
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(params))
	if err != nil {
		fmt.Println("Error initializing genre map:", err)
		return
	}
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Client-ID", apiKey)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error initializing genre map:", err)
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error initializing genre map:", err)
		return
	}

	// Unmarshal the JSON response
	var genres []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	err = json.Unmarshal(body, &genres)
	if err != nil {
		fmt.Println("error unmarshall genres struct(utility.go ligne 235, func initGenreMap)", err)
	}
	if err != nil {
		fmt.Println("Error initializing genre map:", err)
		return
	}

	// Populate the genre map
	for _, genre := range genres {
		GenreMap[genre.ID] = genre.Name
	}

}

func Init() {
	genreMapOnce.Do(initGenreMap)
}

// GetGenreNameByID retrieves the genre name for a given genre ID
func GetGenreNameByID(genreID int) string {
	return GenreMap[genreID]
}

func getCoverImageURL(gameID int) string {
	apiKey := "3pfgrdttfa66z525wc6d40uzjv9nq3" // the client ID
	apiURL := "https://api.igdb.com/v4/covers"
	accessToken, err := GetTwitchOAuthToken()
	if err != nil {
		fmt.Println("error token", err)
		return ""
	}
	if accessToken == "" {
		fmt.Println("no token was given back")
	}

	// Build the parameters for the API request
	params := `
    fields url;
    where id =  ` + strconv.Itoa(gameID) + `;
    `

	// Make a POST request to the IGDB API with the parameters in the body
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(params))
	if err != nil {
		fmt.Println(err)
		return ""
	}

	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Client-ID", apiKey)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	// Unmarshal the JSON response
	var covers []struct {
		URL string `json:"url"`
	}
	err = json.Unmarshal(body, &covers)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	// Extract cover image URL
	if len(covers) > 0 {
		largeCoverURL := strings.Replace(covers[0].URL, "t_thumb", "t_cover_big", 1)
		return "https:" + largeCoverURL
	}
	fmt.Println("no cover found")
	return ""
}

func formatUnixTimestampToFrenchDate(timestamp int64) string {
	// Convert Unix timestamp to time.Time
	utcTime := time.Unix(timestamp, 0)

	// Set the time zone to French time (Europe/Paris)
	frenchLocation, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		fmt.Println("Error loading time zone:", err)
		return ""
	}

	// Convert UTC time to French time
	frenchTime := utcTime.In(frenchLocation)

	// Format the time as DD/MM/YYYY
	formattedDate := frenchTime.Format("02/01/2006")

	return formattedDate
}

func fetchLastGames() ([]Game, error) {
	apiKey := "3pfgrdttfa66z525wc6d40uzjv9nq3" // the client ID
	apiURL := "https://api.igdb.com/v4/games"
	accessToken, err := GetTwitchOAuthToken()
	// keep in mind that max number of token at same time is 25 , oldest one is deleted after
	if err != nil {
		fmt.Println("error token", err)
		return nil, err
	}
	if accessToken == "" {
		fmt.Println("no token was given back")
	}

	// Get the current time in Unix time format
	currentTime := time.Now().Unix()

	// Convert Unix time to a string
	timeString := fmt.Sprintf("%d", currentTime)

	// theme 42 MUST be out , it's erotica theme
	params := `
	fields *;
	where themes != 42 & first_release_date <` + timeString + ` ;
	limit 6;
	sort first_release_date desc;
	`

	// Make a POST request to the IGDB API with the parameters in the body
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(params))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Client-ID", apiKey)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON response
	var games []Game
	err = json.Unmarshal(body, &games)
	if err != nil {
		return nil, err
	}

	// Populate struct for each game
	for i := range games {
		games[i].GenresString = make([]string, len(games[i].Genres))
		for j, genreID := range games[i].Genres {
			games[i].GenresString[j] = GetGenreNameByID(genreID)
		}
	}
	// Populate struct for each game
	for y := range games {
		games[y].CoverLink = getCoverImageURL(games[y].Cover)
		if len(games[y].Screenshots) > 0 {
			games[y].ScreenshotsLink = getScreenshotsImageURL(games[y].Screenshots[0])
			if games[y].ScreenshotsLink == "" {
				games[y].ScreenshotsLink = "static/img/Picture_Not_Yet_Available.png"
			}
		} else {
			games[y].ScreenshotsLink = getScreenshotsImageURL(games[y].Cover)
			if games[y].ScreenshotsLink == "" {
				games[y].ScreenshotsLink = "static/img/Picture_Not_Yet_Available.png"
			}
		}

		games[y].FirstReleaseDateHuman = formatUnixTimestampToFrenchDate(games[y].FirstReleaseDate)
	}

	return games, nil
}

func getScreenshotsImageURL(gameID int) string { // copy past with light modification for screenshots
	apiKey := "3pfgrdttfa66z525wc6d40uzjv9nq3" // the client ID
	apiURL := "https://api.igdb.com/v4/screenshots"
	accessToken, err := GetTwitchOAuthToken()
	if err != nil {
		fmt.Println("error token", err)
		return ""
	}
	if accessToken == "" {
		fmt.Println("no token was given back")
	}

	// Build the parameters for the API request
	params := `
    fields url;
    where id =  ` + strconv.Itoa(gameID) + `;
    `

	// Make a POST request to the IGDB API with the parameters in the body
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(params))
	if err != nil {
		fmt.Println(err)
		return ""
	}

	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Client-ID", apiKey)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	// Unmarshal the JSON response
	var covers []struct {
		URL string `json:"url"`
	}
	err = json.Unmarshal(body, &covers)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	// Extract screenshots image URL
	if len(covers) > 0 {
		largeCoverURL := strings.Replace(covers[0].URL, "t_thumb", "t_cover_big", 1)
		return "https:" + largeCoverURL
	}
	fmt.Println("no screenshots found")
	return ""

}

func fetchGame(id string) GameFull {
	apiKey := "3pfgrdttfa66z525wc6d40uzjv9nq3" // the client ID
	apiURL := "https://api.igdb.com/v4/games"
	accessToken, err := GetTwitchOAuthToken()
	// keep in mind that max number of token at same time is 25 , oldest one is deleted after
	if err != nil {
		fmt.Println("error token", err)
		return GameFull{}
	}
	if accessToken == "" {
		fmt.Println("no token was given back")
		return GameFull{}
	}

	// theme 42 MUST be out , it's erotica theme
	params := `
	fields *;
	where id =` + id

	// Make a POST request to the IGDB API with the parameters in the body
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(params))
	if err != nil {
		fmt.Println("error new request fetchgame", err)
		return GameFull{}
	}

	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Client-ID", apiKey)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("error do req", err)
		return GameFull{}
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error readAll", err)
		return GameFull{}
	}

	// Unmarshal the JSON response
	// Unmarshal the JSON response
	var game GameFull
	err = json.Unmarshal(body, &game)
	if err != nil {
		fmt.Println("error Unmarshal", err)
		return GameFull{}
	}
	fmt.Println(game)
	game.GenresString = make([]string, len(game.Genres))
	for j, genreID := range game.Genres {
		game.GenresString[j] = GetGenreNameByID(genreID)
	}
	game.CoverLink = getCoverImageURL(game.Cover)
	if game.ScreenshotsLink == "" {
		game.ScreenshotsLink = "static/img/Picture_Not_Yet_Available.png"
	}
	game.ScreenshotsLink = getScreenshotsImageURL(game.Cover)
	if game.ScreenshotsLink == "" {
		game.ScreenshotsLink = "static/img/Picture_Not_Yet_Available.png"
	}
	return game
}
