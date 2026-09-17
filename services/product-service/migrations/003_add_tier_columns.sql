-- Add missing columns to product_tiers
ALTER TABLE product_tiers ADD COLUMN IF NOT EXISTS tier_name VARCHAR(255);
ALTER TABLE product_tiers ADD COLUMN IF NOT EXISTS price_usd DECIMAL(12,2) NOT NULL DEFAULT 0;
ALTER TABLE product_tiers ADD COLUMN IF NOT EXISTS file_path TEXT DEFAULT '';
ALTER TABLE product_tiers ADD COLUMN IF NOT EXISTS download_count INTEGER DEFAULT 0;
ALTER TABLE product_tiers ADD COLUMN IF NOT EXISTS download_limit INTEGER DEFAULT 0;
ALTER TABLE product_tiers ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT true;
ALTER TABLE product_tiers ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE;

-- Update existing tier names from the 'name' column if it exists
UPDATE product_tiers SET tier_name = name WHERE tier_name IS NULL;
