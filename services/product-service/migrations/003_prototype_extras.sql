-- Extra columns and product_requests for the live prototype catalog.

ALTER TABLE product_tiers ADD COLUMN IF NOT EXISTS tier_name VARCHAR(255) DEFAULT '';
ALTER TABLE product_tiers ADD COLUMN IF NOT EXISTS price_usd DECIMAL(12,2) DEFAULT 0;
ALTER TABLE product_tiers ADD COLUMN IF NOT EXISTS download_count INTEGER DEFAULT 0;
ALTER TABLE product_tiers ADD COLUMN IF NOT EXISTS download_limit INTEGER DEFAULT 0;
ALTER TABLE product_tiers ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true;
ALTER TABLE product_tiers ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();
UPDATE product_tiers SET tier_name = name WHERE (tier_name IS NULL OR tier_name = '') AND name IS NOT NULL;

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
