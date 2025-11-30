package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func StartOutboxWorker(ctx context.Context, db *sql.DB, rdb *redis.Client) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			processBatch(ctx, db, rdb)
		}
	}
}

func processBatch(ctx context.Context, db *sql.DB, rdb *redis.Client) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Worker failed to begin tx: %v", err)
		return
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `
		SELECT id, event_type, payload
		FROM outbox
		WHERE status = 'PENDING'
		ORDER BY created_at ASC
		LIMIT 10
		FOR UPDATE SKIP LOCKED
	`)
	if err != nil {
		log.Printf("Worker query failed: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		var eventType string
		var payload []byte

		if err := rows.Scan(&id, &eventType, &payload); err != nil {
			continue
		}

		err := rdb.Publish(ctx, eventType, payload).Err()
		if err != nil {
			log.Printf("Failed to publish event %s: %v", id, err)
			return
		}

		_, err = tx.ExecContext(ctx,
			"UPDATE outbox SET status = 'PROCESSED' WHERE id = $1", id)
		if err != nil {
			return
		}
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Worker failed to commit: %v", err)
	}
}
