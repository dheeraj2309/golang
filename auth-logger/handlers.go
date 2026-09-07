package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// register handler-> POST request
// sending the user details
func RegisterHandler(s *UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		// decode the request body
		var creds ClientCredentials
		err := json.NewDecoder(r.Body).Decode(&creds)
		if err != nil {
			http.Error(w, "Invalid JSON Payload", http.StatusBadRequest)
			return
		}
		// check if the data is empty
		if creds.Username == "" || creds.Password == "" {
			http.Error(w, "Username and password cannot be empty", http.StatusBadRequest)
			return
		}
		// hash the received password
		hashedBytes, err := bcrypt.GenerateFromPassword([]byte(creds.Password), 10)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		// now we have the relevant data , so we can create the user record to add to the db
		newUser := User{
			Username:     creds.Username,
			PasswordHash: string(hashedBytes),
		}
		createdUser, err := s.CreateUser(newUser)
		if err != nil {
			if errors.Is(err, ErrUserExists) {
				http.Error(w, "User already exists", http.StatusConflict)
				return
			}
			http.Error(w, "Failed to create new user", http.StatusInternalServerError)
			return
		}
		// to make a structed response we create anonymous struct
		createdResponse := struct {
			ID        int       `json:"id"`
			Username  string    `json:"username"`
			CreatedAt time.Time `json:"createdAt"`
		}{
			ID:        createdUser.ID,
			Username:  createdUser.Username,
			CreatedAt: createdUser.CreatedAt,
		}
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(createdResponse)
	}
}

// GET ->request
// login functionality
func LoginHandler(s *UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var creds ClientCredentials
		err := json.NewDecoder(r.Body).Decode(&creds)
		if err != nil {
			http.Error(w, "Invalid JSON Payload", http.StatusBadRequest)
			return
		}
		//validate the data
		if creds.Username == "" || creds.Password == "" {
			http.Error(w, "Username and password cannot be empty", http.StatusBadRequest)
			return
		}
		// fetch the user
		getUser, err := s.GetUser(creds.Username)
		if err != nil {
			// Security Best Practice: Use a generic error message
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}
		// we got the user, now we have to check the password
		err = bcrypt.CompareHashAndPassword([]byte(getUser.PasswordHash), []byte(creds.Password))
		if err != nil {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}
		// 5. Generate the Authentication Token
		// In a production app, you would generate a cryptographically signed JWT here.
		// For our middleware testing purposes, we will issue a simple mock string token.
		mockToken := "secret-token-for-" + getUser.Username

		// 6. Send the success response
		// We use a map to quickly format the JSON without declaring a dedicated struct
		response := map[string]string{
			"token": mockToken,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

