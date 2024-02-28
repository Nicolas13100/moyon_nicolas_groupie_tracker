package API

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

var (
	users    = make(map[string]User) // Map to store users
	username string
	password string
	logged   bool
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "Register", logged)
}

func confirmRegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")

	// Check if username already exists
	if _, exists := users[username]; exists {
		http.Error(w, "Username already exists", http.StatusConflict)
		return
	}

	hashedPassword := hashPassword(password)
	users[username] = User{Username: username, Password: hashedPassword}

	// Save users to a file
	if err := saveUsersToFile("users.json"); err != nil {
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	// Load users from a file on startup
	if err := loadUsersFromFile("users.json"); err != nil {
		panic(err)
	}
	// Check if there are query parameters in the URL
	queryParams := r.URL.Query()
	// Get a specific query parameter value by key
	invalidParam := queryParams.Get("invalid")
	var Invalid string
	Invalid = ""
	// Use the obtained query parameter value
	if invalidParam != "" {
		Invalid = "Invalid username or password"
		invalidParam = ""
	}

	data := CombinedData{
		Logged: logged,
		Name:   Invalid,
	}

	renderTemplate(w, "Login", data)
}

func successLoginHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	username = r.FormValue("username")
	password = r.FormValue("password")

	user, exists := users[username]
	if !exists || !checkPasswordHash(password, user.Password) {
		http.Redirect(w, r, "/login?invalid=true", http.StatusSeeOther)
		return
	}
	logged = true
	// Successfully logged in
	// Handle further operations (e.g., setting session, redirecting, etc.)
	http.Redirect(w, r, "/home", http.StatusSeeOther)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	ResetUserValue()
	http.Redirect(w, r, "/home", http.StatusSeeOther)
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	if !logged {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	autoData := struct {
		Name   string
		Logged bool
	}{
		Name:   username,
		Logged: logged,
	}
	renderTemplate(w, "dashboard", autoData)
}

func gestionHandler(w http.ResponseWriter, r *http.Request) {
	if !logged {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	data := struct {
		PlayerName string
		Logged     bool
	}{
		PlayerName: username,
		Logged:     logged,
	}
	renderTemplate(w, "gestion", data)
}

func changeLoginHandler(w http.ResponseWriter, r *http.Request) {
	if !logged {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	oldpassword := r.FormValue("oldpassword")
	newpassword := r.FormValue("newpassword")
	err := updateUserCredentials(username, oldpassword, newpassword)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Password updated successfully.")
	ResetUserValue()
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Function to save users to a file for register func
func saveUsersToFile(filename string) error {
	data, err := json.Marshal(users)
	if err != nil {
		return err
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		log.Println("Error writing updated user data:", err)
		return err
	}

	log.Println("User data successfully updated")
	return nil
}

func hashPassword(password string) string {
	hasher := sha256.New()
	hasher.Write([]byte(password))
	hashedPassword := hasher.Sum(nil)
	return hex.EncodeToString(hashedPassword)
}

func checkPasswordHash(password, hash string) bool {
	hashedPassword := hashPassword(password)
	return hashedPassword == hash
}

// Function to load users from a file for register func
func loadUsersFromFile(filename string) error {
	// Check if the file exists
	fileInfo, err := os.Stat(filename)
	if os.IsNotExist(err) {
		// Create an empty users.json file if it doesn't exist
		file, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer file.Close()
	} else if err != nil {
		return err
	}

	// Check if the file is empty
	if fileInfo != nil && fileInfo.Size() == 0 {
		// File exists but is empty, so initialize users as an empty map
		users = make(map[string]User)
		return nil
	}

	// Load users from the file
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	// Check if the file contains valid JSON data
	if len(data) == 0 {
		// File is empty or doesn't contain valid JSON
		return nil
	}

	users = make(map[string]User)
	if err := json.Unmarshal(data, &users); err != nil {
		return err
	}

	return nil
}

func ResetUserValue() {
	logged = false
	username = ""
	password = ""
}

func updateUserCredentials(name, oldPassword, newPassword string) error {
	// Read the JSON file into memory
	raw, err := os.ReadFile("users.json")
	if err != nil {
		return err
	}

	// Define a struct that matches your JSON structure
	var data map[string]User // Map where keys are strings and values are User structs

	// Unmarshal the JSON into the defined struct
	if err := json.Unmarshal(raw, &data); err != nil {
		return err
	}

	// Check if the user exists in the map
	user, exists := data[name]
	if !exists {
		return fmt.Errorf("user not found")
	}

	if !checkPasswordHash(oldPassword, user.Password) {
		return fmt.Errorf("incorrect password")
	}

	if newPassword != "" {
		// Update the password
		user.Password = hashPassword(newPassword)

		// Update the user in the map
		data[name] = user

		// Marshal the updated data back to JSON
		updatedJSON, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return err
		}

		// Write the updated JSON back to the file
		err = os.WriteFile("users.json", updatedJSON, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}
