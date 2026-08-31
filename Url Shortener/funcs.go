package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// `GET /{code}`
func (s *URLStore) getURL(w http.ResponseWriter, r *http.Request) {
	// get the code from the URL
	code := r.PathValue("code")
	if !isalphanum(code) {
		http.Error(w, `{"error":"invalid code format"}`, http.StatusBadRequest) // 400 bad request
		return
	}
	//accessing db
	s.mu.Lock()
	item, ok := s.urls[code]
	if ok {
		item.Hits++
		s.urls[code] = item
	}
	s.mu.Unlock()
	if !ok {
		http.Error(w, `{"error":"short url not found"}`, http.StatusNotFound)
		return
	}
	http.Redirect(w, r, item.OrgURL, http.StatusFound)
}

// `POST /shorten`
func (s *URLStore) shorten(w http.ResponseWriter, r *http.Request) {
	var req ShortenRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, `{"error":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}
	target := strings.TrimSpace(req.URL)
	if target == "" {
		http.Error(w, `{"error":"url is required"}`, http.StatusBadRequest)
		return
	}
	if !isValidURL(target) {
		http.Error(w, `{"error":"invalid url: must start with http:// or https:// and include a host"}`, http.StatusBadRequest)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	const mxAttempt = 5
	var code string
	var erro error
	attempt := 0
	created := false
	// since the random may cause collision to prevent it , we do 5 tries to generate Random code
	for attempt < mxAttempt {
		attempt++
		code, erro = generateRandomCode(6)
		if erro != nil {
			http.Error(w, `{"error":"failed to generate random code"}`, http.StatusInternalServerError)
			return
		}
		if _, ok := s.urls[code]; !ok {
			item := URLItem{
				Code:      code,
				OrgURL:    target,
				Hits:      0,
				CreatedAt: time.Now(),
			}
			s.urls[code] = item
			created = true
			break
		}
	}
	if !created {
		http.Error(w, `{"error":"failed to generate unique code, try again"}`, http.StatusInternalServerError)
		return
	}
	res := ShortenResponse{
		Code:     code,
		ShortURL: fmt.Sprintf("http://localhost:8080/%s", code), // <- Returns a string
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}
