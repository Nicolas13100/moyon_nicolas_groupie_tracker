package API

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"sync"
)

// RUN function sets up HTTP routes and starts the server
func RUN() {
	// Setting up HTTP routes for different endpoints
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
	http.HandleFunc("/favPage", favPageHandler)
	http.HandleFunc("/categorie", categorieHandler)
	http.HandleFunc("/categories", categoriesHandler)

	// Serve static files from the "site_web/static" directory
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("site_web/static"))))

	// Print statement indicating server is running
	fmt.Println("Server is running on :8080 http://localhost:8080/home")

	// Start the server
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

// indexHandler handles requests for the home page
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

// gameHandler handles requests for individual game pages
func gameHandler(w http.ResponseWriter, r *http.Request) {
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
	data, err := fetchGame(idStr)
	if err != nil {
		fmt.Println("gameHandler", err)
	}

	dataS := CombinedData{
		Result: data,
		Logged: logged,
		Fav:    isFav(id),
	}
	renderTemplate(w, "game", dataS)
}

// ErrorHandler handles 404 errors
func ErrorHandler(w http.ResponseWriter, r *http.Request) {

	renderTemplate(w, "404", logged)
}

// favHandler handles adding games to favorites
func favHandler(w http.ResponseWriter, r *http.Request) {
	if !logged {
		// Redirect to login if not logged in
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	idStr := r.FormValue("id")
	fav := r.FormValue("fav")
	if idStr == "" {
		fmt.Println("ID parameter is missing or empty on call to gameHandler")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("Failed to parse ID parameter as integer:", err)
		return
	}
	// if id is error return without saving
	if id == 0 {
		http.Redirect(w, r, "/game?id="+idStr, http.StatusFound)
		return
	}

	if username != "" {
		err = SaveUserFavorit(username, id)
	}

	if err != nil {
		fmt.Println("Failed to save fav :", err)
		return
	}
	if fav == "fav" {
		http.Redirect(w, r, "/favPage", http.StatusFound)
	} else {
		// Redirect to /game with the ID as query parameter
		http.Redirect(w, r, "/game?id="+idStr, http.StatusFound)
	}
}

// searchHandler handles game search functionality
func searchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page <= 0 {
		page = 1
	}
	// Parse the form data
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	// Get the selected tags
	tagStrings := r.Form["tags[]"]

	// Convert tag strings to integers
	var tags []int
	for _, tagString := range tagStrings {
		tagInt, err := strconv.Atoi(tagString)
		if err != nil {
			// Handle error if conversion fails
			http.Error(w, "Invalid tag: "+tagString, http.StatusBadRequest)
			return
		}
		tags = append(tags, tagInt)
	}

	// Process the selected tags
	fmt.Printf("Selected Tags: %v\n", tags)

	searchResults, totalResults, err := fetchSearch(query, page, 5, tags)
	if err != nil {
		fmt.Println(err)
	}

	// Calculate total number of pages
	totalPages := int(math.Ceil(float64(totalResults) / float64(5)))
	// Create a slice containing page numbers from 1 to totalPages
	var pages []int
	for i := 1; i <= totalPages; i++ {
		pages = append(pages, i)
	}
	dataS := CombinedData{
		SearchResult: searchResults,
		Logged:       logged,
		Pagination: PaginationInfo{
			PrevPage:   page - 1,
			Page:       page,
			NextPage:   page + 1,
			TotalPages: totalPages,
			Query:      query,
			Pages:      pages,
		},
	}
	renderTemplate(w, "search", dataS)
}

// favPageHandler renders the user's favorite games page
func favPageHandler(w http.ResponseWriter, r *http.Request) {
	if !logged {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	jsonFileName := username + ".json"

	// Read the JSON data from the file
	jsonData, err := os.ReadFile(jsonFileName)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// Unmarshal the JSON data into a UserData struct
	var userData UserData
	err = json.Unmarshal(jsonData, &userData)
	if err != nil {
		fmt.Println("Error favPage:", err)
		return
	}

	// Extract the IDs from the Fav field
	ids := userData.Fav

	// Create a channel to synchronize goroutines
	done := make(chan bool)

	// Create a wait group to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Create a slice to store the fetched game data
	var games []GameFull

	// Convert each ID to a string and fetch game data concurrently
	for _, id := range ids {
		wg.Add(1) // Increment the wait group counter
		go func(id int) {
			defer wg.Done() // Decrement the wait group counter when done
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("Goroutine panicked with error: %v\n", r)
					return
				}
			}()
			stringID := strconv.Itoa(id)
			data, err := fetchGame(stringID)
			if err != nil {
				fmt.Println("gameHandler", err)
				return
			}

			// Append the fetched data to the slice
			games = append(games, data)
		}(id)
	}

	// Wait for all goroutines to finish
	go func() {
		wg.Wait()
		close(done) // Close the channel to signal completion
	}()

	// Wait for the completion signal
	<-done

	// Now 'games' contains all the fetched game data
	dataS := CombinedData{
		Result: games,
		Logged: logged,
	}
	renderTemplate(w, "fav", dataS)
}

// categorieHandler handles requests to display games by a specific category
func categorieHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("id")
	data := fetchGameByGenres(query)
	dataS := CombinedData{
		Result2: data,
		Logged:  logged,
	}
	renderTemplate(w, "categorie", dataS)
}

// categoriesHandler handles requests to display all categories
func categoriesHandler(w http.ResponseWriter, r *http.Request) {
	data := fetchAllGames()
	dataS := CombinedData{
		Result: data,
		Logged: logged,
	}
	renderTemplate(w, "categories", dataS)
}
