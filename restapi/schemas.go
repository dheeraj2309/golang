package main

import (
	"sync"
	"time"
)

type Task struct {
	ID          int        `json:"id"`
	Title       string     `json:"title"`
	Done        bool       `json:"done"`
	CreatedAt   time.Time  `json:"createdAt"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
}

type CreateTaskRequest struct {
	Title string `json:"title"`
}

type TaskStore struct {
	mu     sync.Mutex
	task   []Task
	nextID int
}

func NewTaskStore() *TaskStore {
	return &TaskStore{
		task:   []Task{},
		nextID: 1,
	}
}
