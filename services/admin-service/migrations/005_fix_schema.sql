-- Fix missing columns in admin DB tables

-- Products: add asset_path, asset_hash
ALTER TABLE products ADD COLUMN IF NOT EXISTS asset_path TEXT DEFAULT '';
ALTER TABLE products ADD COLUMN IF NOT EXISTS asset_hash TEXT DEFAULT '';

-- Coupons: add missing columns
ALTER TABLE coupons ADD COLUMN IF NOT EXISTS expires_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE coupons ADD COLUMN IF NOT EXISTS usage_limit INTEGER DEFAULT 0;
ALTER TABLE coupons ADD COLUMN IF NOT EXISTS times_used INTEGER DEFAULT 0;
ALTER TABLE coupons ADD COLUMN IF NOT EXISTS min_purchase_usd NUMERIC(12,2) DEFAULT 0;
ALTER TABLE coupons ADD COLUMN IF NOT EXISTS product_id INTEGER;
ALTER TABLE coupons ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true;

-- Settings: add setting_key column (rename from key if needed)
ALTER TABLE settings ADD COLUMN IF NOT EXISTS setting_key VARCHAR(255);

-- Migrate existing data: copy key -> setting_key if setting_key is null
UPDATE settings SET setting_key = key WHERE setting_key IS NULL;

-- Community posts: add missing columns
ALTER TABLE community_posts ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();

-- Referral links: add is_active column
ALTER TABLE referral_links ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true;

-- Orders: add guest_email column for guest orders
ALTER TABLE orders ADD COLUMN IF NOT EXISTS guest_email VARCHAR(255);
