package API

import (
	"fmt"
	"log"
	"net/http"
)

func RUN() {
	// used same system than hangman, since it was working prety well
	http.HandleFunc("/", indexHandler)

	// Serve static files from the "site_web/static" directory << modified from hangman
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("../site_web/static"))))

	// Print statement indicating server is running << same
	fmt.Println("Server is running on :8080 http://localhost:8080")

	// Start the server << same
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "index", nil)
}
