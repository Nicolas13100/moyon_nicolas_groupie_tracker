package API

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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
		games[y].CoverLink = getSavedCoverImageURL(games[y].Cover)
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
	savedData := map[string]string{
		"0": "static/img/Picture_Not_Yet_Available.png",
	}
	if _, err := os.Stat("savedCover.json"); os.IsNotExist(err) {
		file, err := os.Create("savedCover.json")
		if err != nil {
			fmt.Println("Error creating savedCover.json, please ensure it's properly here for quick page loading")
		}
		defer file.Close()
		// Encode the data and write it to the file
		encoder := json.NewEncoder(file)
		err = encoder.Encode(savedData)
		if err != nil {
			fmt.Println("Error encoding data:", err)
			return
		}
		fmt.Println("savedCover.json created and initialized successfully")
	} else {
		fmt.Println("savedCover.json already exists")
	}
	if _, err := os.Stat("savedScreenShot.json"); os.IsNotExist(err) {
		file, err := os.Create("savedScreenShot.json")
		if err != nil {
			fmt.Println("Error creating savedScreenShot.json, please ensure it's properly here for quick page loading")
		}
		defer file.Close()
		// Encode the data and write it to the file
		encoder := json.NewEncoder(file)
		err = encoder.Encode(savedData)
		if err != nil {
			fmt.Println("Error encoding data:", err)
			return
		}
		fmt.Println("savedScreenShot.json created and initialized successfully")
	} else {
		fmt.Println("savedScreenShot.json already exists")
	}
}

// GetGenreNameByID retrieves the genre name for a given genre ID
func GetGenreNameByID(genreID int) string {
	return GenreMap[genreID]
}

func getCoverImageURL(coverID int) string {
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
    where id =  ` + strconv.Itoa(coverID) + `;
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
		largeCoverURL := strings.Replace(covers[0].URL, "t_thumb", "t_cover_big_2x", 1)
		coverLink := "https:" + largeCoverURL
		err := saveCoverImageURL(coverID, coverLink)
		if err != nil {
			fmt.Println("error saving Cover image", err)
		}
		return coverLink
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
		games[y].CoverLink = getSavedCoverImageURL(games[y].Cover)
		if len(games[y].Screenshots) > 0 {
			games[y].ScreenshotsLink = getSavedScreenShotImageURL(games[y].Screenshots[0])
			if games[y].ScreenshotsLink == "" {
				games[y].ScreenshotsLink = "static/img/Picture_Not_Yet_Available.png"
			}
		} else {
			games[y].ScreenshotsLink = getSavedScreenShotImageURL(games[y].Cover)
			if games[y].ScreenshotsLink == "" {
				games[y].ScreenshotsLink = "static/img/Picture_Not_Yet_Available.png"
			}
		}

		games[y].FirstReleaseDateHuman = formatUnixTimestampToFrenchDate(games[y].FirstReleaseDate)
	}

	return games, nil
}

func getScreenshotsImageURL(ScreenshotsID int) string { // copy past with light modification for screenshots

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
    where id =  ` + strconv.Itoa(ScreenshotsID) + `;
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
		largeCoverURL := strings.Replace(covers[0].URL, "t_thumb", "t_1080p", 1)
		screenshotsLink := "https:" + largeCoverURL
		err := saveScreenShotsImageURL(ScreenshotsID, screenshotsLink)
		if err != nil {
			fmt.Println("error saving ScreenShots image", err)
		}
		return screenshotsLink
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
	where id =` + id + `;`

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
	var game []GameFull
	err = json.Unmarshal(body, &game)
	if err != nil {
		fmt.Println("error Unmarshal", err)
		return GameFull{}
	}
	// Populate struct for each game
	for i := range game {
		game[i].GenresString = make([]string, len(game[i].Genres))
		for j, genreID := range game[i].Genres {
			game[i].GenresString[j] = GetGenreNameByID(genreID)
		}
	}
	// Populate struct for each game
	for y := range game {
		game[y].CoverLink = getSavedCoverImageURL(game[y].Cover)
		if len(game[y].Screenshots) > 0 {
			game[y].ScreenshotsLink = getScreenshotsImageURL(game[y].Screenshots[0])
			if game[y].ScreenshotsLink == "" {
				game[y].ScreenshotsLink = "static/img/Picture_Not_Yet_Available.png"
			}
		} else {
			game[y].ScreenshotsLink = getScreenshotsImageURL(game[y].Cover)
			if game[y].ScreenshotsLink == "" {
				game[y].ScreenshotsLink = "static/img/Picture_Not_Yet_Available.png"
			}
		}

		game[y].FirstReleaseDateHuman = formatUnixTimestampToFrenchDate(game[y].FirstReleaseDate)
	}
	return game[0]
}

func getSavedCoverImageURL(coverID int) string {
	// Open the saved.json file
	file, err := os.Open("savedCover.json")
	if err != nil {
		fmt.Println("error opening savedCover.json:", err)
		return ""
	}
	defer file.Close()

	// Decode the JSON data
	var savedData map[string]string
	if err := json.NewDecoder(file).Decode(&savedData); err != nil {
		fmt.Println("error decoding savedCover.json:", err)
		return ""
	}

	// Check if the gameID exists in the saved data
	if coverURL, ok := savedData[strconv.Itoa(coverID)]; ok {
		// If the coverURL is not the default placeholder URL, return it
		if coverURL != "static/img/Picture_Not_Yet_Available.png" {
			return coverURL
		}
	}

	// If no matching cover URL found in saved data, continue with getCoverImageURL function
	return getCoverImageURL(coverID)
}

func saveCoverImageURL(coverID int, coverURL string) error {
	// Open the JSON file for reading and writing
	file, err := os.OpenFile("savedCover.json", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Decode the existing JSON data into a map
	var data map[int]string
	err = json.NewDecoder(file).Decode(&data)
	if err != nil && err != io.EOF {
		return err
	}

	// Check if the cover ID already exists in the map
	if _, ok := data[coverID]; ok {
		// If the cover ID exists, replace the cover URL
		data[coverID] = coverURL
	} else {
		// If the cover ID doesn't exist, add it to the map
		data[coverID] = coverURL
	}

	// Move the file cursor to the beginning
	_, err = file.Seek(0, 0)
	if err != nil {
		return err
	}

	// Truncate the file (remove any extra data)
	err = file.Truncate(0)
	if err != nil {
		return err
	}

	// Encode the updated map and write it back to the file
	err = json.NewEncoder(file).Encode(data)
	if err != nil {
		return err
	}

	return nil
}

func getSavedScreenShotImageURL(ScreenshotsID int) string {
	// Open the saved.json file
	file, err := os.Open("savedScreenShot.json")
	if err != nil {
		fmt.Println("error opening savedScreenShot.json:", err)
		return ""
	}
	defer file.Close()

	// Decode the JSON data
	var savedData map[string]string
	if err := json.NewDecoder(file).Decode(&savedData); err != nil {
		fmt.Println("error decoding savedScreenShot.json:", err)
		return ""
	}

	// Check if the gameID exists in the saved data
	if coverURL, ok := savedData[strconv.Itoa(ScreenshotsID)]; ok {
		// If the coverURL is not the default placeholder URL, return it
		if coverURL != "static/img/Picture_Not_Yet_Available.png" {
			return coverURL
		}
	}

	// If no matching cover URL found in saved data, continue with getCoverImageURL function
	return getScreenshotsImageURL(ScreenshotsID)
}

func saveScreenShotsImageURL(ScreenshotsID int, ScreenshotsURL string) error {
	// Open the JSON file for reading and writing
	file, err := os.OpenFile("savedScreenShot.json", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Decode the existing JSON data into a map
	var data map[int]string
	err = json.NewDecoder(file).Decode(&data)
	if err != nil && err != io.EOF {
		return err
	}

	// Check if the Screenshots ID already exists in the map
	if _, ok := data[ScreenshotsID]; ok {
		// If the Screenshots ID exists, replace the Screenshots URL
		data[ScreenshotsID] = ScreenshotsURL
	} else {
		// If the Screenshots ID doesn't exist, add it to the map
		data[ScreenshotsID] = ScreenshotsURL
	}

	// Move the file cursor to the beginning
	_, err = file.Seek(0, 0)
	if err != nil {
		return err
	}

	// Truncate the file (remove any extra data)
	err = file.Truncate(0)
	if err != nil {
		return err
	}

	// Encode the updated map and write it back to the file
	err = json.NewEncoder(file).Encode(data)
	if err != nil {
		return err
	}

	return nil
}
