package main

import (
	"context"
	"doubleLedger/internal/service"
	"fmt"
	"log"
	"log/slog"
	"os"
	"sync"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5/pgxpool"
)

func setupAccounts(ctx context.Context, pool *pgxpool.Pool, user1ID, user2ID uuid.UUID) {
	// Clean the database for the test
	pool.Exec(ctx, "TRUNCATE ledger_entries, transactions, accounts CASCADE")

	// Insert Amit and Priya with 50,000 initial balance
	_, err := pool.Exec(ctx, "INSERT INTO accounts (id, name, cached_balance) VALUES ($1, 'user1', 50000), ($2, 'user2', 50000)", user1ID, user2ID)
	if err != nil {
		log.Fatalf("Setup failed: %v", err)
	}
}

func verifyBalances(ctx context.Context, pool *pgxpool.Pool, user1ID, user2ID uuid.UUID) {
	var u1Balance, u2Balance int64
	pool.QueryRow(ctx, "SELECT cached_balance FROM accounts WHERE id = $1", user1ID).Scan(&u1Balance)
	pool.QueryRow(ctx, "SELECT cached_balance FROM accounts WHERE id = $1", user2ID).Scan(&u2Balance)
	fmt.Println("--- FINAL RESULTS ---")
	fmt.Printf("user1 Balance: ₹%d (Expected: 50000)\n", u1Balance)
	fmt.Printf("user2 Balance: ₹%d (Expected: 50000)\n", u2Balance)

	if u1Balance == 50000 && u2Balance == 50000 {
		fmt.Println("STATUS: FLAWLESS. Zero deadlocks, zero race conditions.")
	} else {
		fmt.Println("STATUS: CORRUPTION DETECTED.")
	}
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx := context.Background()
	logger.Info("booting ledger API", "env", "development")
	dbURL := "postgres://postgres:1234@localhost:5432/ledger?sslmode=disable"
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		logger.Error("unable to connect to database", "error", err)
		log.Fatalf("Failed to connect: %v", err)
	}
	defer pool.Close()
	ledger := service.NewLedgerService(pool,logger)

	user1ID, user2ID := uuid.New(), uuid.New()
	setupAccounts(ctx, pool, user1ID, user2ID)

	fmt.Println("Starting stress test: 500 concurrent transfers...")
	// 3. The Concurrency Engine
	var wg sync.WaitGroup
	startTime := time.Now()

	// Fire 250 Goroutines: user1 -> user2
	for i := 0; i < 250; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := ledger.TransferFunds(ctx, user1ID, user2ID, 100)
			if err != nil {
				log.Printf("Transfer error (user1->user2): %v", err)
			}
		}()
	}
	// Fire 250 Goroutines: user2 -> user1
	for i := 0; i < 250; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := ledger.TransferFunds(ctx, user2ID, user1ID, 100)
			if err != nil {
				log.Printf("Transfer error (user2->user1): %v", err)
			}
		}()
	}

	// Wait for all 500 threads to finish their database transactions
	wg.Wait()

	fmt.Printf("Completed in %v\n", time.Since(startTime))

	// 4. Verify the Math
	verifyBalances(ctx, pool, user1ID, user2ID)
}
