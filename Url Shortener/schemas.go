package main

import (
	"sync"
	"time"
)

type URLItem struct {
	Code      string    `json:"code"`
	OrgURL    string    `json:"original_url"`
	CreatedAt time.Time `json:"created_at"`
	Hits      int       `json:"hits"` // tracks how many times redirected
}
type URLStore struct {
	mu   sync.RWMutex
	urls map[string]URLItem // shortcode -> orgURL
}

func NewURLStore() *URLStore {
	return &URLStore{
		urls: make(map[string]URLItem),
	}
}

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	Code     string `json:"code"`
	ShortURL string `json:"shorturl"`
}
