CREATE TABLE IF NOT EXISTS payments (
    payment_id UUID PRIMARY KEY NOT NULL,
    order_id UUID NOT NULL,
    amount INTEGER NOT NULL,
    payment_method VARCHAR(50),
    status VARCHAR(50),
    created_at TIMESTAMP DEFAULT NOW()
);