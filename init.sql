-- 1. The Business Table: Stores the actual state
CREATE TABLE orders (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    amount INT NOT NULL, 
    status VARCHAR(50) DEFAULT 'PENDING',
    created_at TIMESTAMP DEFAULT NOW()
);

-- 2. The Outbox Table: Stores the intent to publish
CREATE TABLE outbox (
    id UUID PRIMARY KEY,
    event_type VARCHAR(255) NOT NULL, -- e.g., "order.created"
    payload JSONB NOT NULL,           -- The actual data to publish
    status VARCHAR(50) DEFAULT 'PENDING',
    created_at TIMESTAMP DEFAULT NOW()
);