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

		fmt.Println("Failed to fetch recommended games", err)
		return
	}
	// Fetch last added games from IGDB API
	lastGame, err := fetchLastGames()
	if err != nil {

		fmt.Println("Failed to fetch recommended games", err)
		return
	}
	data := TemplateData{
		RecommendedGames: recommendedGames,
		LastGames:        lastGame,
	}
	// Render the index template with the data
	renderTemplate(w, "index", data)
}

func ErrorHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "404", nil)
}
