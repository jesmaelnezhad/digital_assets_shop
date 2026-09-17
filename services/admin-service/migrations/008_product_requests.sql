CREATE TABLE IF NOT EXISTS product_requests (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT DEFAULT '',
    category TEXT DEFAULT '',
    budget_usd DECIMAL(12,2) DEFAULT 0,
    email VARCHAR(255) DEFAULT '',
    status VARCHAR(50) DEFAULT 'open',
    user_id INTEGER,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
