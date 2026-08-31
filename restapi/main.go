package main

import (
	"log/slog"
	"net/http"
	"os"
)

func main() {
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		// If the file can't be created, print error to stderr and exit
		os.Stderr.WriteString("failed to open log file: " + err.Error() + "\n")
		os.Exit(1)
	}
	defer logFile.Close()
	logger := slog.New(slog.NewJSONHandler(logFile,&slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)
	store := NewTaskStore()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		slog.Info("PING HIT")
		w.Write([]byte("pong"))
	})
	// register "GET /tasks" to the mux router
	mux.HandleFunc("GET /tasks", store.getTasks)

	// register `POST /tasks`
	mux.HandleFunc("POST /tasks", store.insertTask)
	// register `GET /tasks/{id}`
	mux.HandleFunc("GET /tasks/{id}", store.getTaskByID)
	// register `PATCH /tasks/{id}/toggle`
	mux.HandleFunc("PATCH /tasks/{id}/toggle", store.toggleTask)
	// register `DELETE /tasks/{id}`
	mux.HandleFunc("DELETE /tasks/{id}", store.deleteTask)
	loggedMux := loggingMiddleware(mux)
	addr := "127.0.0.1:8080"
	slog.Info("server starting", "addr", addr)
	if err := http.ListenAndServe(addr, loggedMux); err != nil {
		slog.Error("server failed to start", "error", err)
		os.Exit(1)
	}
}
