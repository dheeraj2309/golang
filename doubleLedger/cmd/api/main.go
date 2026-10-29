package api

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbURL := "postgres://postgres:1234@localhost:5432/postgres?sslmode=disable"
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v\n", err)
	}
	defer pool.Close()
	err = pool.Ping(context.Background())
	if err != nil {
		log.Fatalf("Database didn't respond: %v\n", err)
	}
	fmt.Println("Fortress connected successfully.")
}
