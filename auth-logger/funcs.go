package main

import (
	"errors"
	"time"
)

var (
	ErrUserExists   = errors.New("User already exists")
	ErrUserNotFound = errors.New("User not found")
)

// creates the user in the database
func (s *UserStore) CreateUser(u User) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock() // no matter how this function ends , unlock the db
	if user, ok := s.users[u.Username]; ok {
		return user, ErrUserExists
	}
	u.ID = s.nextID
	u.CreatedAt = time.Now()
	s.nextID++
	s.users[u.Username] = u
	return u, nil
}

// we have to fetch the user
func (s *UserStore) GetUser(username string) (User, error) {
	s.mu.RLock()
	user, ok := s.users[username]
	s.mu.RUnlock()
	if !ok {
		return User{}, ErrUserNotFound
	} else {
		return user, nil
	}
}
