-- identity volume
-- appdb_identity_staging
-- Password hash matches demo nia so collector-001@example.com / nia works.
-- =============================================================================

INSERT INTO users (email, password_hash, name)
SELECT 'collector-' || lpad(g::text, 3, '0') || '@example.com',
  '$2b$10$blu/StVHqw2O.hZO7HrYEuETFYgw8bICkm1QMChmFVU.i.bT94AZa',
  'Collector ' || g
FROM generate_series(1, 80) g
ON CONFLICT (email) DO NOTHING;

INSERT INTO user_profiles (user_id, display_name, bio, avatar_url)
SELECT u.id, u.name, 'Volume collector #' || u.id, '/assets/catalog/p0' || ((u.id % 9) + 1) || '.jpg'
FROM users u
WHERE u.email LIKE 'collector-%@example.com'
ON CONFLICT (user_id) DO UPDATE SET bio = EXCLUDED.bio;

-- =============================================================================
