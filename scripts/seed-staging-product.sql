-- Apply to appdb_product_staging and appdb_admin_staging.

INSERT INTO categories (name, slug, description, sort_order, is_active) VALUES
('3D & characters', '3d', 'Sculpts, kits, and game-ready meshes', 1, true),
('UI kits', 'ui', 'Dashboards, components, tokens', 2, true),
('Textures', 'textures', 'Materials, HDRI, surfaces', 3, true),
('Audio', 'audio', 'Loops, kits, foley', 4, true),
('Environments', 'environments', 'Scenes and world kits', 5, true),
('Icons & type', 'icons', 'Icon sets and specimens', 6, true),
('VFX', 'vfx', 'Sprites, volumes, sequences', 7, true)
ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name, description = EXCLUDED.description, is_active = true;

INSERT INTO products (title, slug, description, category_id, price_usd, status, image_url, pinned, sort_order, digital_formats, tags, is_pwyw, pwyw_min_price, purchase_count, views)
SELECT 'Lunar Clay Characters', 'lunar-clay-characters', 'Twelve stylized clay figures, game-ready, 8k texture set, and a small turntable scene.', c.id, 48, 'active', '/assets/catalog/p01.jpg', true, 1, 'FBX', 'clay,character', false, 0, 910, 910
FROM categories c WHERE c.slug = '3d'
ON CONFLICT (slug) DO UPDATE SET title = EXCLUDED.title, description = EXCLUDED.description, price_usd = EXCLUDED.price_usd, image_url = EXCLUDED.image_url, pinned = EXCLUDED.pinned, digital_formats = EXCLUDED.digital_formats;

INSERT INTO products (title, slug, description, category_id, price_usd, status, image_url, pinned, sort_order, digital_formats, tags, is_pwyw, pwyw_min_price, purchase_count)
SELECT 'Northline UI Kit', 'northline-ui-kit', '80 frames of dense product UI: tables, filters, empty states. Dark and light.', c.id, 36, 'active', '/assets/catalog/p02.jpg', true, 2, 'Figma', 'ui', false, 0, 640
FROM categories c WHERE c.slug = 'ui'
ON CONFLICT (slug) DO UPDATE SET title = EXCLUDED.title, image_url = EXCLUDED.image_url, pinned = EXCLUDED.pinned, digital_formats = EXCLUDED.digital_formats;

INSERT INTO products (title, slug, description, category_id, price_usd, status, image_url, pinned, sort_order, digital_formats, tags, purchase_count)
SELECT 'Obsidian Marble Pack', 'obsidian-marble-pack', 'Fourteen seamless marble and stone materials with displacement and roughness.', c.id, 19, 'active', '/assets/catalog/p03.jpg', false, 3, 'PNG', 'marble', 1200
FROM categories c WHERE c.slug = 'textures'
ON CONFLICT (slug) DO UPDATE SET title = EXCLUDED.title, image_url = EXCLUDED.image_url, digital_formats = EXCLUDED.digital_formats;

INSERT INTO products (title, slug, description, category_id, price_usd, status, image_url, pinned, sort_order, digital_formats, tags, purchase_count)
SELECT 'Night Bus Lo-fi Kit', 'night-bus-lofi', 'Tape-worn keys, bus interiors, rain on glass. 48 stems, tempo-labeled.', c.id, 15, 'active', '/assets/catalog/p04.jpg', false, 4, 'WAV', 'audio', 330
FROM categories c WHERE c.slug = 'audio'
ON CONFLICT (slug) DO UPDATE SET title = EXCLUDED.title, image_url = EXCLUDED.image_url, digital_formats = EXCLUDED.digital_formats;

INSERT INTO products (title, slug, description, category_id, price_usd, status, image_url, pinned, sort_order, digital_formats, tags, purchase_count)
SELECT 'Concrete Atlas', 'concrete-atlas', 'A brutalist courtyard and tower. Modular walls, hero camera, dusk HDRI.', c.id, 64, 'active', '/assets/catalog/p05.jpg', false, 5, 'USD', 'environment', 210
FROM categories c WHERE c.slug = 'environments'
ON CONFLICT (slug) DO UPDATE SET title = EXCLUDED.title, image_url = EXCLUDED.image_url, digital_formats = EXCLUDED.digital_formats;

INSERT INTO products (title, slug, description, category_id, price_usd, status, image_url, pinned, sort_order, digital_formats, tags, purchase_count)
SELECT 'Glyph Factory Icons', 'glyph-factory', '420 stroke icons on a 1.5px optical grid. Figma + SVG.', c.id, 12, 'active', '/assets/catalog/p06.jpg', false, 6, 'SVG', 'icons', 1500
FROM categories c WHERE c.slug = 'icons'
ON CONFLICT (slug) DO UPDATE SET title = EXCLUDED.title, image_url = EXCLUDED.image_url, digital_formats = EXCLUDED.digital_formats;

INSERT INTO products (title, slug, description, category_id, price_usd, status, image_url, pinned, sort_order, digital_formats, tags, purchase_count)
SELECT 'Drift Type Specimen', 'drift-specimen', 'A display serif with a slightly broken italic. Two optical sizes.', c.id, 29, 'active', '/assets/catalog/p07.jpg', false, 7, 'OTF', 'type', 90
FROM categories c WHERE c.slug = 'icons'
ON CONFLICT (slug) DO UPDATE SET title = EXCLUDED.title, image_url = EXCLUDED.image_url, digital_formats = EXCLUDED.digital_formats;

INSERT INTO products (title, slug, description, category_id, price_usd, status, image_url, pinned, sort_order, digital_formats, tags, purchase_count)
SELECT 'Canopy Environment', 'canopy-environment', 'Dappled forest floor, wind-ready foliage, and a path that leads somewhere.', c.id, 42, 'active', '/assets/catalog/p08.jpg', false, 8, 'USD', 'environment', 400
FROM categories c WHERE c.slug = 'environments'
ON CONFLICT (slug) DO UPDATE SET title = EXCLUDED.title, image_url = EXCLUDED.image_url, digital_formats = EXCLUDED.digital_formats;

INSERT INTO products (title, slug, description, category_id, price_usd, status, image_url, pinned, sort_order, digital_formats, tags, purchase_count)
SELECT 'Signal VFX Pack', 'signal-vfx', 'Lens dirt, hologram tiles, energy bursts. 32-bit EXR sequences.', c.id, 27, 'active', '/assets/catalog/p09.jpg', false, 9, 'EXR', 'vfx', 140
FROM categories c WHERE c.slug = 'vfx'
ON CONFLICT (slug) DO UPDATE SET title = EXCLUDED.title, image_url = EXCLUDED.image_url, digital_formats = EXCLUDED.digital_formats;

INSERT INTO products (title, slug, description, category_id, price_usd, status, image_url, pinned, sort_order, digital_formats, tags, purchase_count)
SELECT 'Arcade Tile Kit', 'arcade-tile-kit', '16×16 dungeon and shop tiles with a matching character sheet.', c.id, 9, 'active', '/assets/catalog/p10.jpg', false, 10, 'PNG', 'pixel', 2200
FROM categories c WHERE c.slug = '3d'
ON CONFLICT (slug) DO UPDATE SET title = EXCLUDED.title, image_url = EXCLUDED.image_url, digital_formats = EXCLUDED.digital_formats;

INSERT INTO products (title, slug, description, category_id, price_usd, status, image_url, pinned, sort_order, digital_formats, tags, purchase_count)
SELECT 'Studio Light HDRIs', 'studio-light-hdris', 'Six studio wraps, from hard beauty to dusty warehouse.', c.id, 22, 'active', '/assets/catalog/p11.jpg', false, 11, 'HDR', 'hdri', 500
FROM categories c WHERE c.slug = 'textures'
ON CONFLICT (slug) DO UPDATE SET title = EXCLUDED.title, image_url = EXCLUDED.image_url, digital_formats = EXCLUDED.digital_formats;

INSERT INTO products (title, slug, description, category_id, price_usd, status, image_url, pinned, sort_order, digital_formats, tags, is_pwyw, pwyw_min_price, purchase_count)
SELECT 'Sketchbook Brushes', 'sketch-brushes', 'Pay what you want, floor $4. Graphite and wash brushes.', c.id, 8, 'active', '/assets/catalog/p12.jpg', false, 12, 'ABR', 'brushes', true, 4, 70
FROM categories c WHERE c.slug = 'ui'
ON CONFLICT (slug) DO UPDATE SET title = EXCLUDED.title, image_url = EXCLUDED.image_url, is_pwyw = true, pwyw_min_price = 4, digital_formats = EXCLUDED.digital_formats;

INSERT INTO bundles (title, slug, description, price_usd, status, sort_order) VALUES
('Studio Kit', 'studio-kit', 'Clay characters, marble, and HDRIs — the stills stack.', 72, 'active', 1),
('Jam Pack', 'jam-pack', 'Tiles, icons, and lo-fi for a weekend ship.', 28, 'active', 2)
ON CONFLICT (slug) DO UPDATE SET title = EXCLUDED.title, price_usd = EXCLUDED.price_usd;
