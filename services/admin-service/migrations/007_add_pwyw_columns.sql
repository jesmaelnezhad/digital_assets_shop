-- Add missing columns to admin DB products table

ALTER TABLE products ADD COLUMN IF NOT EXISTS pwyw_enabled BOOLEAN DEFAULT false;
ALTER TABLE products ADD COLUMN IF NOT EXISTS pwyw_min_price NUMERIC(12,2) DEFAULT 0;
