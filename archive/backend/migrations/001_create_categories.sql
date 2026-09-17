-- 001_create_categories.sql
-- Categories for digital assets

CREATE TABLE IF NOT EXISTS categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    parent_id INTEGER REFERENCES categories(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_categories_slug ON categories(slug);

-- Seed default categories
INSERT INTO categories (name, slug, description) VALUES
    ('Graphics & Design', 'graphics-design', 'Icons, logos, illustrations, templates, fonts'),
    ('Code & Scripts', 'code-scripts', 'Source code, scripts, plugins, themes, libraries'),
    ('3D Models & Assets', '3d-models', '3D models, textures, animations, STL files'),
    ('Audio & Music', 'audio-music', 'Sound effects, music tracks, samples, loops'),
    ('Documents & Templates', 'documents-templates', 'PDFs, spreadsheets, presentations, contracts'),
    ('Courses & Tutorials', 'courses-tutorials', 'Learning materials, video courses, guides'),
    ('Games & Game Assets', 'games-game-assets', 'Game mods, assets, sprites, sound packs'),
    ('Other Digital Goods', 'other', 'Anything that doesn\'t fit the above categories')
ON CONFLICT (slug) DO NOTHING;
