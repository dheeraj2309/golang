package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TransferRequest struct {
	FromAccountID uuid.UUID `json:"from_account_id"`
	ToAccountID   uuid.UUID `json:"to_account_id"`
	Amount        int64     `json:"amount"`
}

var (
	successCount int32
	blockedCount int32

	// Mutex and slices to safely capture samples across 200+ Goroutines
	sampleMu       sync.Mutex
	successSamples []string
	rejectSamples  []string
)

func main() {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, "postgres://postgres:1234@localhost:5432/ledger?sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer pool.Close()

	_, err = pool.Exec(ctx, "TRUNCATE ledger, transactions, accounts CASCADE")
	if err != nil {
		log.Fatalf("Failed to truncate: %v", err)
	}

	user1ID, user2ID := uuid.New(), uuid.New()
	pool.Exec(ctx, "INSERT INTO accounts (id, name, cached_balance) VALUES ($1, 'user1', 50000), ($2, 'user2', 50000)", user1ID, user2ID)

	fmt.Println("==================================================")
	fmt.Printf(" [DB] user1 UUID: %s (Balance: ₹50,000)\n", user1ID)
	fmt.Printf(" [DB] user2 UUID: %s (Balance: ₹50,000)\n", user2ID)
	fmt.Println("==================================================")
	fmt.Println("\nInitializing Concurrent Network Attack...")
	time.Sleep(2 * time.Second)

	var wg sync.WaitGroup
	client := &http.Client{Timeout: 10 * time.Second}

	go func() {
		for {
			s := atomic.LoadInt32(&successCount)
			b := atomic.LoadInt32(&blockedCount)
			total := s + b
			fmt.Printf("\r Processing: %d Valid | %d Rejected | %d/215 Total", s, b, total)
			if total >= 215 {
				fmt.Println("\n\nTraffic Simulation Complete.")
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	for i := 0; i < 100; i++ {
		wg.Add(1)
		time.Sleep(30 * time.Millisecond)
		go sendRequest(&wg, client, user1ID, user2ID, 100, true)
	}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		time.Sleep(30 * time.Millisecond)
		go sendRequest(&wg, client, user2ID, user1ID, 100, true)
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		time.Sleep(100 * time.Millisecond)
		go sendRequest(&wg, client, user1ID, user2ID, 9999999, true)
	}

	for i := 0; i < 5; i++ {
		wg.Add(1)
		time.Sleep(100 * time.Millisecond)
		go sendRequest(&wg, client, user1ID, user2ID, 50, false)
	}

	wg.Wait()
	time.Sleep(1 * time.Second)

	var user1Bal, user2Bal int64
	pool.QueryRow(ctx, "SELECT cached_balance FROM accounts WHERE id = $1", user1ID).Scan(&user1Bal)
	pool.QueryRow(ctx, "SELECT cached_balance FROM accounts WHERE id = $1", user2ID).Scan(&user2Bal)

	fmt.Println("==================================================")
	fmt.Println("              FINAL DATABASE STATE                ")
	fmt.Println("==================================================")
	fmt.Printf(" user1 Balance: ₹%d (Expected: 50000)\n", user1Bal)
	fmt.Printf(" user2 Balance: ₹%d (Expected: 50000)\n", user2Bal)

	if user1Bal == 50000 && user2Bal == 50000 {
		fmt.Println("\n STATUS: FLAWLESS. Zero race conditions detected.")
	} else {
		fmt.Println("\n STATUS: CORRUPTION. Engine failed.")
	}

	// Print the collected samples
	fmt.Println("\n==================================================")
	fmt.Println("            API RESPONSE SAMPLES                  ")
	fmt.Println("==================================================")
	fmt.Println(" [ACCEPTED REQUESTS]")
	for _, s := range successSamples {
		fmt.Println(s)
	}
	fmt.Println("\n [REJECTED REQUESTS]")
	for _, s := range rejectSamples {
		fmt.Println(s)
	}
	fmt.Println("==================================================")
}

func sendRequest(wg *sync.WaitGroup, client *http.Client, from, to uuid.UUID, amount int64, includeKey bool) {
	defer wg.Done()

	payload := TransferRequest{
		FromAccountID: from,
		ToAccountID:   to,
		Amount:        amount,
	}

	jsonData, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "http://127.0.0.1:8080/transfer", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	if includeKey {
		req.Header.Set("X-Idempotency-Key", uuid.New().String())
	}

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	
	// Read the response body to capture error messages
	bodyBytes, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	respBody := strings.TrimSpace(string(bodyBytes))

	// Safely update counters and capture samples
	sampleMu.Lock()
	defer sampleMu.Unlock()

	shortFrom := from.String()
	shortTo := to.String()

	if resp.StatusCode == http.StatusOK {
		atomic.AddInt32(&successCount, 1)
		if len(successSamples) < 5 {
			successSamples = append(successSamples, fmt.Sprintf(" 200 OK | %s -> %s | ₹%d", shortFrom, shortTo, amount))
		}
	} else {
		atomic.AddInt32(&blockedCount, 1)
		if len(rejectSamples) < 5 {
			rejectSamples = append(rejectSamples, fmt.Sprintf(" %d | %s -> %s | ₹%d | Reason: %s", resp.StatusCode, shortFrom, shortTo, amount, respBody))
		}
	}
}