package main

import (
	"log/slog"
	"net/http"
	"os"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		slog.SetDefault(logger)
	
		store := NewURLStore()
		mx := http.NewServeMux()
	
		// Register routes
		mx.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
			slog.Info("PING HIT")
			w.Write([]byte("pong"))
		})
		mx.HandleFunc("GET /{code}", store.getURL)
		mx.HandleFunc("POST /shorten", store.shorten)
	
		addr := "127.0.0.1:8080"
		slog.Info("URL Shortener service starting", "addr", addr)
	
		// ListenAndServe always returns a non-nil error when it stops
		if err := http.ListenAndServe(addr, mx); err != nil {
			slog.Error("server failed to start", "error", err)
			os.Exit(1)
		}
}