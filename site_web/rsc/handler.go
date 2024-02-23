package API

import (
	"fmt"
	"log"
	"net/http"
	"sync"
)

func RUN() {
	// used same system than hangman, since it was working prety well
	http.HandleFunc("/", ErrorHandler)
	http.HandleFunc("/home", indexHandler)
	http.HandleFunc("/game", gameHandler)

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
	// Once data is fetched, use JavaScript to update the DOM and remove the loading indicator
	fmt.Fprintf(w, "<script>document.getElementById('loading').remove();</script>")

	// Render the index template with the data
	data := TemplateData{
		RecommendedGames: recommendedGames,
		LastGames:        lastGame,
	}
	renderTemplate(w, "index", data)
}

func gameHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the ID from the query parameters
	// Extract the ID from the query parameters
	id := r.FormValue("id")
	if id == "" {
		fmt.Println("ID parameter is missing or empty on call to gameHandler")
		return
	}
	data := fetchGame(id)

	renderTemplate(w, "game", data)
}

func ErrorHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "404", nil)
}
