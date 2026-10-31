package main

import (
	"context"
	"doubleLedger/internal/service"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5/pgxpool"
)


type TransferRequest struct{
	FromAccountID uuid.UUID `json:"from_account_id"`
	ToAccountID uuid.UUID `json:"to_account_id"`
	Amount int64 `json:"amount"`
}

func main(){
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx := context.Background()
	logger.Info("booting ledger API", "env", "development")
	pool,err := pgxpool.New(ctx,"postgres://postgres:1234@localhost:5432/ledger?sslmode=disable")
	if err != nil{
		logger.Error("unable to connect to database", "error", err)
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()
	ledgerSvc := service.NewLedgerService(pool,logger)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /transfer",func(w http.ResponseWriter,r *http.Request){
		
		idempotencyKey := r.Header.Get("X-Idempotency-Key")
		if idempotencyKey == "" {
			logger.Warn("rejected: missing idempotency key", slog.String("ip", r.RemoteAddr))
			http.Error(w, "Missing X-Idempotency-Key header", http.StatusBadRequest)
			return
		}
		
		// Decode the JSON payload
		var req TransferRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}
		err := ledgerSvc.TransferFunds(r.Context(), req.FromAccountID, req.ToAccountID, req.Amount)
		if err != nil {
			// In production, you'd want to separate 400 Bad Request (Insufficient funds) 
			// from 500 Internal Server Error (Database down)
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		// Send success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "success", "message": "Transfer completed"}`))
	})

	// 4. Start the Server
	server := &http.Server{
		Addr:         "127.0.0.1:8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	fmt.Println("Ledger API running on port 8080...")
	log.Fatal(server.ListenAndServe())
}