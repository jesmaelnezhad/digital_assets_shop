-- product + admin catalog volume
-- appdb_product_staging  (also safe on appdb_admin_staging for catalog copies)
-- =============================================================================

INSERT INTO categories (name, slug, description, sort_order, is_active) VALUES
('Motion', 'motion', 'Loops, titles, and boards', 8, true),
('Print', 'print', 'Posters, editorial, layouts', 9, true),
('Photos', 'photos', 'Stills and plates', 10, true)
ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name, description = EXCLUDED.description, is_active = true;

WITH bases AS (
  SELECT p.*, ROW_NUMBER() OVER (ORDER BY p.id) AS rn
  FROM products p
  WHERE p.slug IN (
    'lunar-clay-characters','northline-ui-kit','obsidian-marble-pack','night-bus-lofi',
    'concrete-atlas','glyph-factory','drift-specimen','canopy-environment',
    'signal-vfx','arcade-tile-kit','studio-light-hdris','sketch-brushes'
  )
)
INSERT INTO products (title, slug, description, category_id, price_usd, status, image_url, pinned, sort_order, digital_formats, tags, purchase_count)
SELECT
  b.title || ' Vol. ' || g.n,
  b.slug || '-vol-' || g.n,
  b.description,
  COALESCE((SELECT id FROM categories WHERE slug = (ARRAY['3d','ui','textures','audio','environments','icons','vfx','motion','print','photos'])[ ((g.n - 1) % 10) + 1 ]), b.category_id),
  GREATEST(6, b.price_usd + ((g.n % 7) * 3)),
  'active',
  b.image_url,
  false,
  100 + g.n,
  b.digital_formats,
  COALESCE(b.tags, 'volume'),
  40 + g.n
FROM generate_series(1, 36) g(n)
JOIN bases b ON b.rn = ((g.n - 1) % 12) + 1
ON CONFLICT (slug) DO NOTHING;

UPDATE products SET banner_sort = 0 WHERE COALESCE(banner_sort, 0) <> 0;
UPDATE products SET pinned = false WHERE pinned = true AND slug NOT IN ('lunar-clay-characters','northline-ui-kit','concrete-atlas');
UPDATE products SET banner_sort = 1, pinned = true WHERE slug = 'lunar-clay-characters';
UPDATE products SET banner_sort = 2, pinned = true WHERE slug = 'northline-ui-kit';
UPDATE products SET banner_sort = 3, pinned = true WHERE slug = 'concrete-atlas';

INSERT INTO bundles (title, slug, description, price_usd, status, sort_order) VALUES
('Lookdev Stack', 'lookdev-stack', 'Courtyard, canopy, and marble for stills.', 96, 'active', 3),
('Weekend HUD', 'weekend-hud', 'UI kit, icons, and type for a HUD jam.', 58, 'active', 4),
('Forest Set', 'forest-set', 'Canopy plus VFX plates.', 54, 'active', 5)
ON CONFLICT (slug) DO UPDATE SET title = EXCLUDED.title, price_usd = EXCLUDED.price_usd, description = EXCLUDED.description;

INSERT INTO bundle_items (bundle_id, product_id, sort_order)
SELECT b.id, p.id, x.sort_order
FROM (VALUES
  ('lookdev-stack', 'concrete-atlas', 1),
  ('lookdev-stack', 'canopy-environment', 2),
  ('lookdev-stack', 'obsidian-marble-pack', 3),
  ('weekend-hud', 'northline-ui-kit', 1),
  ('weekend-hud', 'glyph-factory', 2),
  ('weekend-hud', 'drift-specimen', 3),
  ('forest-set', 'canopy-environment', 1),
  ('forest-set', 'signal-vfx', 2)
) AS x(bundle_slug, product_slug, sort_order)
JOIN bundles b ON b.slug = x.bundle_slug
JOIN products p ON p.slug = x.product_slug
WHERE NOT EXISTS (
  SELECT 1 FROM bundle_items bi WHERE bi.bundle_id = b.id AND bi.product_id = p.id
);

-- =============================================================================
