-- 003_create_product_images.sql
-- Product image references

CREATE TABLE IF NOT EXISTS product_images (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    url VARCHAR(500) NOT NULL,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_product_images_product ON product_images(product_id);

-- Seed sample images (using placeholder service)
INSERT INTO product_images (product_id, url, is_primary)
SELECT id, 'https://placehold.co/600x400/0a0a0f/00d4ff?text=Pixel+Icon+Pack', TRUE FROM products WHERE slug = 'pixel-icon-pack-500'
ON CONFLICT DO NOTHING;

INSERT INTO product_images (product_id, url, is_primary)
SELECT id, 'https://placehold.co/600x400/0a0a0f/00d4ff?text=Dashboard+Template', TRUE FROM products WHERE slug = 'vanilla-js-dashboard'
ON CONFLICT DO NOTHING;

INSERT INTO product_images (product_id, url, is_primary)
SELECT id, 'https://placehold.co/600x400/0a0a0f/ff0066?text=Low-Poly+Trees', TRUE FROM products WHERE slug = 'low-poly-tree-pack-30'
ON CONFLICT DO NOTHING;

INSERT INTO product_images (product_id, url, is_primary)
SELECT id, 'https://placehold.co/600x400/0a0a0f/00ff88?text=Synthwave+Sounds', TRUE FROM products WHERE slug = 'synthwave-sound-pack-1'
ON CONFLICT DO NOTHING;

INSERT INTO product_images (product_id, url, is_primary)
SELECT id, 'https://placehold.co/600x400/0a0a0f/00d4ff?text=README+Template', TRUE FROM products WHERE slug = 'ultimate-readme-template'
ON CONFLICT DO NOTHING;
