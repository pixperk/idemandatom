package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type Order struct {
	ID     uuid.UUID `json:"id"`
	UserID uuid.UUID `json:"user_id"`
	Amount int       `json:"amount"`
}

func CreateOrder(ctx context.Context, db *sql.DB, order Order) error {
	// 1. Start the Transaction
	// This is the boundary of our atomicity.
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Defer a rollback. If the function exits without a Commit,
	// all changes (both order and outbox) are discarded.
	defer tx.Rollback()

	// 2. Insert the Business Record
	_, err = tx.ExecContext(ctx,
		`INSERT INTO orders (id, user_id, amount) VALUES ($1, $2, $3)`,
		order.ID, order.UserID, order.Amount)
	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	// 3. Insert the Outbox Record
	// We marshal the order data to JSON to serve as the event payload.
	payload, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO outbox (id, event_type, payload) VALUES ($1, $2, $3)`,
		uuid.New(), "order.created", payload)
	if err != nil {
		return fmt.Errorf("failed to insert outbox event: %w", err)
	}

	// 4. Commit the Transaction
	// This is the moment of truth. Both the order and the event become visible
	// to the rest of the system at the exact same instant.
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func ConnectDB() (*sql.DB, error) {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "outbox_db")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	for i := 0; i < 5; i++ {
		if err := db.Ping(); err == nil {
			log.Println("Connected to database successfully")
			return db, nil
		}
		log.Printf("Waiting for database... attempt %d/5", i+1)
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("failed to connect to database after retries")
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func main() {
	db, err := ConnectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	log.Println("Application started")
}
