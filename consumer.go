package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/redis/go-redis/v9"
)

func StartEmailConsumer(ctx context.Context, rdb *redis.Client) {
	sub := rdb.Subscribe(ctx, "order.created")
	defer sub.Close()

	log.Println("Email consumer subscribed to 'order.created' channel")

	ch := sub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-ch:
			handleOrderCreated(msg.Payload)
		}
	}
}

func handleOrderCreated(payload string) {
	var order Order
	if err := json.Unmarshal([]byte(payload), &order); err != nil {
		log.Printf("[EMAIL] Failed to parse order: %v", err)
		return
	}

	log.Printf("[EMAIL] Sending confirmation email for Order %s", order.ID)
	log.Printf("[EMAIL] -> User: %s", order.UserID)
	log.Printf("[EMAIL] -> Amount: %d", order.Amount)
	log.Printf("[EMAIL] ✓ Email sent successfully!")
}
