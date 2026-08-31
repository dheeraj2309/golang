package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// this function on hitting `GET /tasks` returns all the tasks
func (s *TaskStore) getTasks(w http.ResponseWriter, r *http.Request) {
	// lock the in memory db
	s.mu.Lock()
	defer s.mu.Unlock() // <- delay unlocking it unless the work in this function is done
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(s.task); err != nil {
		slog.Error("failed to encode tasks response", "error", err)
	}
}

// `POST /tasks`
func (s *TaskStore) insertTask(w http.ResponseWriter, r *http.Request) {
	// first we store the value that came in request body
	var req CreateTaskRequest // we only expect the title

	// 1. Stream & decode JSON from the request socket
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Warn("malformed json in create task", "error", err)
		http.Error(w, `{"error": "invalid JSON payload"}`, http.StatusBadRequest) // 400 bad request
		return
	}

	//validate the receieved input (should not be empty)
	desc := strings.TrimSpace(req.Title)
	if desc == "" {
		slog.Warn("create task failed: title is empty")
		http.Error(w, `{"error": "title cannot be empty"}`, http.StatusBadRequest) // 400 Bad Request
		return
	}
	// now we have the valid input , now we talk to db
	s.mu.Lock()
	newTask := Task{
		ID:        s.nextID,
		Title:     desc,
		Done:      false,
		CreatedAt: time.Now(),
	}
	s.task = append(s.task, newTask)
	s.nextID++
	s.mu.Unlock() // why not defering it , we dont want to lock throughout the function , bcz there is not db work during validation and sending
	// ok response, we locked only while modifying the data

	slog.Info("task created successfully", "id", newTask.ID, "title", newTask.Title)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newTask)
}

// `GET /tasks/{id}`
func (s *TaskStore) getTaskByID(w http.ResponseWriter, r *http.Request) {
	// get the taskID from the request body
	strID := r.PathValue("id")
	id, err := strconv.Atoi(strID)
	if err != nil {
		slog.Warn("invalid task ID requested", "raw_id", strID)
		http.Error(w, `{"error": "id must be a valid integer"}`, http.StatusBadRequest) // 400 Bad Request
		return
	}
	// we have valid int ID
	// talk to database
	s.mu.Lock()
	var found bool
	var tcopy Task
	for _, t := range s.task {
		if t.ID == id {
			found = true
			tcopy = t
			break
		}
	}
	s.mu.Unlock()
	if !found {
		slog.Warn("task not found", "id", id)
		http.Error(w, `{"error": "task not found"}`, http.StatusNotFound) // 404 Not Found
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tcopy)
}

// `PATCH /tasks/{id}/toggle`
func (s *TaskStore) toggleTask(w http.ResponseWriter, r *http.Request) {
	// get the ID from request
	strID := r.PathValue("id")
	id, err := strconv.Atoi(strID)
	if err != nil {
		slog.Warn("invalid task ID requested for toggle", "raw_id", strID)
		http.Error(w, `{"error": "id must be a valid integer"}`, http.StatusBadRequest)
		return
	}
	// talk to db
	s.mu.Lock()
	var changedTask Task
	var found bool
	for i := range s.task {
		if s.task[i].ID == id {
			s.task[i].Done = !s.task[i].Done
			if s.task[i].Done {
				now := time.Now()
				s.task[i].CompletedAt = &now
			} else {
				s.task[i].CompletedAt = nil
			}
			found = true
			changedTask = s.task[i]
			break
		}
	}
	s.mu.Unlock()

	if !found {
		slog.Warn("task not found for toggle", "id", id)
		http.Error(w, `{"error": "task not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(changedTask)
}

// `DELETE /tasks/{id}
func (s *TaskStore) deleteTask(w http.ResponseWriter, r *http.Request) {
	strID := r.PathValue("id")
	id, err := strconv.Atoi(strID)
	if err != nil {
		slog.Warn("invalid task ID requested for toggle", "raw_id", strID)
		http.Error(w, `{"error": "id must be a valid integer"}`, http.StatusBadRequest)
		return
	}
	// talk to db
	s.mu.Lock()
	foundIdx := -1
	for i, t := range s.task {
		if t.ID == id {
			foundIdx = i
			break
		}
	}
	if foundIdx == -1 {
		s.mu.Unlock() // Unlock before returning error!
		slog.Warn("task not found for deletion", "id", id)
		http.Error(w, `{"error": "task not found"}`, http.StatusNotFound)
		return
	}
	s.task = append(s.task[:foundIdx], s.task[foundIdx+1:]...) // expects (slice, element1,element2) (...) -> unfolds the slice into element list
	s.mu.Unlock()
	slog.Info("task deleted successfully", "id", id)
	// 3. Respond with 204 No Content
	w.WriteHeader(http.StatusNoContent) //204 No Content is the standard HTTP status code for successful deletions
}
