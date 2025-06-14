-- Create orders table
CREATE TABLE orders
(
    id           UUID PRIMARY KEY,
    user_id      UUID           NOT NULL,
    amount       DECIMAL(15, 2) NOT NULL CHECK (amount >= 0),
    currency     VARCHAR(3)     NOT NULL DEFAULT 'USD',
    status       VARCHAR(50)    NOT NULL DEFAULT 'created' CHECK (status IN ('created', 'payment_pending', 'paid', 'payment_failed', 'completed', 'cancelled')),
    payment_id   VARCHAR(36),
    error_reason TEXT,
    created_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create inbox_messages table for transactional inbox pattern
CREATE TABLE inbox_messages
(
    id           UUID PRIMARY KEY,
    event_id     VARCHAR(36) NOT NULL UNIQUE,
    event_type   VARCHAR(50) NOT NULL,
    payload      JSONB       NOT NULL,
    status       VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processed', 'failed')),
    processed_at TIMESTAMP WITH TIME ZONE,
    created_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    retry_count  INT         NOT NULL DEFAULT 0,
    max_retries  INT         NOT NULL DEFAULT 3
);

-- Outbox table for Transactional Outbox pattern
CREATE TABLE outbox_messages
(
    id          UUID PRIMARY KEY,
    event_type  VARCHAR(50) NOT NULL,
    payload     JSONB       NOT NULL,
    status      VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'sent', 'failed')),
    sent_at     TIMESTAMP WITH TIME ZONE,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    retry_count INT         NOT NULL DEFAULT 0,
    max_retries INT         NOT NULL DEFAULT 3
);

-- Create indexes for better query performance
CREATE INDEX idx_orders_user_id ON orders (user_id);
CREATE INDEX idx_orders_status ON orders (status);
CREATE INDEX idx_orders_created_at ON orders (created_at);
CREATE INDEX idx_orders_payment_id ON orders (payment_id);

-- Inbox indexes for efficient processing
CREATE INDEX idx_inbox_event_id ON inbox_messages (event_id);
CREATE INDEX idx_inbox_status ON inbox_messages (status);
CREATE INDEX idx_inbox_event_type ON inbox_messages (event_type);
CREATE INDEX idx_inbox_created_at ON inbox_messages (created_at);

-- Outbox indexes for efficient processing
CREATE INDEX idx_outbox_status ON outbox_messages (status);
CREATE INDEX idx_outbox_event_type ON outbox_messages (event_type);
CREATE INDEX idx_outbox_created_at ON outbox_messages (created_at);

-- Create function to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers to automatically update updated_at on row updates
CREATE TRIGGER update_orders_updated_at
    BEFORE UPDATE
    ON orders
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_inbox_messages_updated_at
    BEFORE UPDATE
    ON inbox_messages
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_outbox_messages_updated_at
    BEFORE UPDATE
    ON outbox_messages
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
