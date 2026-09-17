-- 001_media.sql
-- Media Service migration: product_images table with extended schema
-- This table stores references to product images, previews, and thumbnails.
-- The media service owns this table for file reference tracking.

CREATE TABLE IF NOT EXISTS product_images (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    url VARCHAR(500) NOT NULL,
    image_type VARCHAR(20) NOT NULL DEFAULT 'full'
        CHECK (image_type IN ('full', 'preview', 'thumbnail')),
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    width INTEGER,
    height INTEGER,
    file_size_bytes BIGINT,
    mime_type VARCHAR(100),
    storage_path VARCHAR(500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_product_images_product ON product_images(product_id);
CREATE INDEX IF NOT EXISTS idx_product_images_type ON product_images(image_type);
CREATE INDEX IF NOT EXISTS idx_product_images_primary ON product_images(product_id, is_primary) WHERE is_primary = TRUE;

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

DROP TRIGGER IF EXISTS update_product_images_updated_at ON product_images;
CREATE TRIGGER update_product_images_updated_at
    BEFORE UPDATE ON product_images
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Down migration (rollback):
-- DROP TRIGGER IF EXISTS update_product_images_updated_at ON product_images;
-- DROP FUNCTION IF EXISTS update_updated_at_column();
-- DROP INDEX IF EXISTS idx_product_images_primary;
-- DROP INDEX IF EXISTS idx_product_images_type;
-- DROP INDEX IF EXISTS idx_product_images_product;
-- DROP TABLE IF EXISTS product_images;
