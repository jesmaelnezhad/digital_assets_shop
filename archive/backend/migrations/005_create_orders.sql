-- 005_create_orders.sql
-- Orders for digital asset purchases

CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(30) NOT NULL DEFAULT 'pending' CHECK (
        status IN ('pending', 'paid', 'failed', 'cancelled', 'refunded')
    ),
    total_usd DECIMAL(10, 2) NOT NULL DEFAULT 0.00,
    total_crypto VARCHAR(50) NOT NULL DEFAULT '0',
    crypto_chain VARCHAR(30) NOT NULL DEFAULT 'bsc',
    payment_address VARCHAR(100) NOT NULL DEFAULT '',
    payment_tx_hash VARCHAR(100),
    payment_confirmations INTEGER NOT NULL DEFAULT 0,
    payment_confirmed_at TIMESTAMPTZ,
    paid_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_user ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created ON orders(created_at);

CREATE TABLE IF NOT EXISTS exchange_rates (
    id SERIAL PRIMARY KEY,
    chain VARCHAR(30) NOT NULL UNIQUE,
    symbol VARCHAR(20) NOT NULL,
    rate_to_usd DECIMAL(20, 8) NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO exchange_rates (chain, symbol, rate_to_usd) VALUES
    ('bsc', 'BNB', 350.00)
ON CONFLICT (chain) DO NOTHING;
