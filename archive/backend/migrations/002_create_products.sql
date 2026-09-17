-- 002_create_products.sql
-- Digital assets for sale

CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    category_id INTEGER REFERENCES categories(id) ON DELETE SET NULL,
    price_usd DECIMAL(10, 2) NOT NULL DEFAULT 0.00,
    asset_path VARCHAR(500) NOT NULL DEFAULT '',
    asset_hash VARCHAR(128) NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'draft', 'archived')),
    download_count_limit INTEGER NOT NULL DEFAULT 0 CHECK (download_count_limit >= 0), -- 0 = unlimited
    max_downloads_per_user INTEGER NOT NULL DEFAULT 0 CHECK (max_downloads_per_user >= 0), -- 0 = unlimited
    file_size_bytes BIGINT,
    file_mime_type VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_products_slug ON products(slug);
CREATE INDEX idx_products_category ON products(category_id);
CREATE INDEX idx_products_status ON products(status);
CREATE INDEX idx_products_price ON products(price_usd);

-- Full-text search
ALTER TABLE products ADD COLUMN search_vector tsvector
    GENERATED ALWAYS AS (to_tsvector('english', title || ' ' || description)) STORED;
CREATE INDEX idx_products_search ON products USING GIN(search_vector);

-- Seed sample products
INSERT INTO products (title, slug, description, category_id, price_usd, status) VALUES
    ('Pixel Icon Pack - 500 Icons', 'pixel-icon-pack-500', 'A comprehensive pack of 500 pixel-perfect icons in PNG and SVG formats. Covers UI elements, common actions, file types, and more. Perfect for web and app design projects.', 1, 12.00, 'active'),
    ('Vanilla JS Dashboard Template', 'vanilla-js-dashboard', 'A clean, dark-themed dashboard template built with vanilla HTML, CSS, and JavaScript. No frameworks, no build step. Includes charts placeholder, data table, sidebar navigation, and responsive layout.', 2, 25.00, 'active'),
    ('Low-Poly Tree Pack - 30 Models', 'low-poly-tree-pack-30', '30 unique low-poly 3D tree models in OBJ format. Various species and sizes. Ready for game engines and rendering. Includes diffuse textures.', 3, 18.00, 'active'),
    ('Synthwave Sound Pack Vol.1', 'synthwave-sound-pack-1', '15 atmospheric synthwave music loops and 25 sound effects. Perfect for game soundtracks, videos, and creative projects. WAV format, 44.1kHz.', 4, 9.00, 'active'),
    ('Ultimate README Template', 'ultimate-readme-template', 'A beautifully crafted README.md template with multiple sections, badges placeholder, screenshot layout, and installation instructions. Markdown format, fully customizable.', 5, 5.00, 'active')
ON CONFLICT (slug) DO NOTHING;
