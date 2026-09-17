-- Add missing columns to admin DB tables

-- Users table already has: id, email, name, created_at, updated_at
-- Add role column
ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(50) DEFAULT 'user';

-- Orders table: add missing columns
ALTER TABLE orders ADD COLUMN IF NOT EXISTS total_crypto DECIMAL(20,8) DEFAULT 0;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS crypto_chain VARCHAR(50) DEFAULT '';
ALTER TABLE orders ADD COLUMN IF NOT EXISTS payment_tx_hash VARCHAR(255) DEFAULT '';
ALTER TABLE orders ADD COLUMN IF NOT EXISTS payment_confirmations INTEGER DEFAULT 0;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS paid_at TIMESTAMP WITH TIME ZONE;

-- Products table: add missing columns
ALTER TABLE products ADD COLUMN IF NOT EXISTS max_downloads_per_user INTEGER DEFAULT 1;
ALTER TABLE products ADD COLUMN IF NOT EXISTS views INTEGER DEFAULT 0;
ALTER TABLE products ADD COLUMN IF NOT EXISTS downloads INTEGER DEFAULT 0;
ALTER TABLE products ADD COLUMN IF NOT EXISTS purchase_count INTEGER DEFAULT 0;
ALTER TABLE products ADD COLUMN IF NOT EXISTS description TEXT DEFAULT '';
ALTER TABLE products ADD COLUMN IF NOT EXISTS category_id INTEGER;

-- Categories table (for admin)
CREATE TABLE IF NOT EXISTS categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    description TEXT DEFAULT '',
    parent_id INTEGER REFERENCES categories(id),
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
