package API

import (
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
	var SearchResult []GameFull
	// Extract the search query from the form
	query := r.URL.Query().Get("query")

	SearchResult = fetchSearch(query)

	dataS := CombinedData{
		Result: SearchResult,
		Logged: logged,
	}

	renderTemplate(w, "search", dataS)
}
