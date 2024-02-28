package API

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
)

func RUN() {
	// used same system than hangman, since it was working prety well
	http.HandleFunc("/", ErrorHandler)
	http.HandleFunc("/home", indexHandler)
	http.HandleFunc("/game", gameHandler)
	http.HandleFunc("/register", RegisterHandler)
	http.HandleFunc("/confirmRegister", confirmRegisterHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/successLogin", successLoginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/dashboard", dashboardHandler)
	http.HandleFunc("/gestion", gestionHandler)
	http.HandleFunc("/changeLogin", changeLoginHandler)
	http.HandleFunc("/fav", favHandler)
	http.HandleFunc("/search", searchHandler)

	// Serve static files from the "site_web/static" directory << modified from hangman
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("site_web/static"))))

	// Print statement indicating server is running << same
	fmt.Println("Server is running on :8080 http://localhost:8080/home")

	// Start the server << same
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	// Use Goroutines to fetch recommended and last added games concurrently
	var recommendedGames []Game
	var lastGame []Game
	var wg sync.WaitGroup
	wg.Add(2)

	// Fetch recommended games concurrently
	go func() {
		defer wg.Done()
		var err error
		recommendedGames, err = fetchRecommendedGames()
		if err != nil {
			fmt.Println("Failed to fetch recommended games", err)
		}
	}()

	// Fetch last added games concurrently
	go func() {
		defer wg.Done()
		var err error
		lastGame, err = fetchLastGames()
		if err != nil {
			fmt.Println("Failed to fetch last added games", err)
		}
	}()

	// Wait for both Goroutines to finish
	wg.Wait()

	// Render the index template with the data
	data := TemplateData{
		RecommendedGames: recommendedGames,
		LastGames:        lastGame,
		Logged:           logged,
	}
	renderTemplate(w, "index", data)
}

func gameHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the ID from the query parameters
	// Extract the ID from the query parameters
	idStr := r.FormValue("id")
	if idStr == "" {
		fmt.Println("ID parameter is missing or empty on call to gameHandler")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("Failed to parse ID parameter as integer:", err)
		return
	}
	data := fetchGame(idStr)
	dataS := CombinedData{
		Result: data,
		Logged: logged,
		Fav:    isFav(id),
	}
	renderTemplate(w, "game", dataS)
}

func ErrorHandler(w http.ResponseWriter, r *http.Request) {

	renderTemplate(w, "404", logged)
}

func favHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.FormValue("id")
	if idStr == "" {
		fmt.Println("ID parameter is missing or empty on call to gameHandler")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("Failed to parse ID parameter as integer:", err)
		return
	}
	err = SaveUserFavorit(username, id)
	if err != nil {
		fmt.Println("Failed to save fav :", err)
		return
	}
	// Redirect to /game with the ID as query parameter
	http.Redirect(w, r, "/game?id="+idStr, http.StatusFound)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	apiKey := "3pfgrdttfa66z525wc6d40uzjv9nq3"
	accessToken, err := GetTwitchOAuthToken()
	// Parse the form data
	if err != nil {
		fmt.Println("error token", err)
		return
	}
	if accessToken == "" {
		fmt.Println("no token was given back")
		return
	}
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	// Extract the search query from the form
	query := r.Form.Get("query")

	// Define the payload
	payload := map[string]interface{}{
		"search": query,
		"fields": "*",
		"limit":  10,
	}

	// Convert payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "Error marshalling JSON", http.StatusInternalServerError)
		return
	}

	// Create a new HTTP request
	req, err := http.NewRequest("POST", "https://api.igdb.com/v4/search", bytes.NewBuffer(payloadBytes))
	if err != nil {
		http.Error(w, "Error creating request", http.StatusInternalServerError)
		return
	}

	// Set headers (including authentication if required)
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Client-ID", apiKey)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	// Create an HTTP client and make the request
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Error making request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Check if the request was successful
	if resp.StatusCode == http.StatusOK {
		// Decode the response JSON
		var result interface{}
		err := json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			http.Error(w, "Error decoding JSON", http.StatusInternalServerError)
			return
		}

		// Write the response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	} else {
		http.Error(w, "Request failed with status: "+resp.Status, resp.StatusCode)
	}
}
