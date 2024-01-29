package API

import (
	"fmt"
	"log"
	"net/http"
)

func RUN() {
	// used same system than hangman, since it was working prety well
	http.HandleFunc("/", ErrorHandler)
	http.HandleFunc("/home", indexHandler)

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
	// Fetch recommended games from IGDB API
	recommendedGames, err := fetchRecommendedGames()
	if err != nil {
		// Handle the error (e.g., log it, return an error response)
		http.Error(w, "Failed to fetch recommended games", http.StatusInternalServerError)
		return
	}

	// Render the index template with the data
	renderTemplate(w, "index", recommendedGames)
}

func ErrorHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "404", nil)
}
