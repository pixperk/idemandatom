# Idempotent Outbox Pattern Demo

Demonstrates the transactional outbox pattern with idempotency protection using Go, PostgreSQL, and Redis.

## Architecture

```
POST /orders (with Idempotency-Key)
       │
       ▼
┌──────────────────┐
│  Lua Script      │ ← Redis: Check/Lock idempotency key
└──────────────────┘
       │
       ▼
┌──────────────────┐
│  PostgreSQL TX   │ ← Atomic: Insert order + outbox event
└──────────────────┘
       │
       ▼
┌──────────────────┐
│  Outbox Worker   │ ← Polls outbox, publishes to Redis
└──────────────────┘
       │
       ▼
┌──────────────────┐
│  Email Consumer  │ ← Subscribes to Redis, "sends" email
└──────────────────┘
```

## Quick Start

### 1. Start Infrastructure

```bash
docker-compose up -d
```

### 2. Run the Application

```bash
go run .
```

### 3. Create an Order

```bash
curl -X POST http://localhost:8080/orders \
  -H "Idempotency-Key: order-123"
```

Response:
```json
{"status": "success", "order_id": "uuid-here"}
```

### 4. Test Idempotency

Send the same request again with the same key:

```bash
curl -X POST http://localhost:8080/orders \
  -H "Idempotency-Key: order-123"
```

You'll get the cached response (no duplicate order created).

## Watch the Logs

You should see:
```
Connected to database successfully
Connected to Redis successfully
Outbox worker started
Email consumer started
Server starting on port 8080
[EMAIL] Sending confirmation email for Order abc-123
[EMAIL] -> User: user-456
[EMAIL] -> Amount: 1000
[EMAIL] ✓ Email sent successfully!
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| DB_HOST | localhost | PostgreSQL host |
| DB_PORT | 5432 | PostgreSQL port |
| DB_USER | postgres | PostgreSQL user |
| DB_PASSWORD | postgres | PostgreSQL password |
| DB_NAME | outbox_db | PostgreSQL database |
| REDIS_ADDR | localhost:6379 | Redis address |
| REDIS_PASSWORD | (empty) | Redis password |
| PORT | 8080 | HTTP server port |

## Load Test

Send 100 concurrent orders:

```bash
go run test_orders.go
```

Output:
```
[0] 200: {"status": "success", "order_id": "..."}
[1] 200: {"status": "success", "order_id": "..."}
...

--- Summary ---
Total: 100 orders
Success: 100
Failed: 0
Duration: 245ms
Rate: 408.16 orders/sec
```

## Files

- `main.go` - HTTP server, handlers, DB/Redis connections
- `worker.go` - Outbox polling worker
- `consumer.go` - Redis pub/sub email consumer
- `script.lua` - Idempotency Lua script for Redis
- `init.sql` - Database schema
- `test_orders.go` - Load test script
