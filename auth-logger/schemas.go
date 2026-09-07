package main

import (
	"sync"
	"time"
)

type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"passwordHash"`
	CreatedAt    time.Time `json:"createdAt"`
}

type ClientCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserStore struct {
	mu    sync.RWMutex
	users map[string]User // helps me to find the user details in O(1) time
	nextID int
}

func NewUserStore() *UserStore {
	return &UserStore{
		users: make(map[string]User),
		nextID: 1,
	}
}
