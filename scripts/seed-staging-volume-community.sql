-- community volume
-- appdb_community_staging
-- Extra directory users 1001–1080 (directory + feed). Demo nia remains 622–625.
-- =============================================================================

INSERT INTO users (id, email, name)
SELECT 1000 + g, 'collector-' || lpad(g::text, 3, '0') || '@example.com', 'Collector ' || g
FROM generate_series(1, 80) g
ON CONFLICT (id) DO UPDATE SET email = EXCLUDED.email, name = EXCLUDED.name;

SELECT setval('users_id_seq', GREATEST((SELECT MAX(id) FROM users), 1080));

INSERT INTO user_profiles (user_id, display_name, bio, avatar_url)
SELECT 1000 + g, 'Collector ' || g, 'Volume collector. Library notes and stills.',
  '/assets/catalog/p0' || ((g % 9) + 1) || '.jpg'
FROM generate_series(1, 80) g
ON CONFLICT (user_id) DO UPDATE SET bio = EXCLUDED.bio, avatar_url = EXCLUDED.avatar_url;

INSERT INTO community_posts (user_id, content, community_type, is_public)
SELECT 1000 + ((g - 1) % 80) + 1,
  'Volume note #' || g || '. Clay, marble, kits, and HDRIs in rotation.',
  'post', true
FROM generate_series(1, 60) g
WHERE NOT EXISTS (
  SELECT 1 FROM community_posts WHERE content = 'Volume note #' || g || '. Clay, marble, kits, and HDRIs in rotation.'
);

INSERT INTO post_comments (post_id, user_id, content)
SELECT p.id, 1000 + ((p.id % 80) + 1), 'Volume comment on post ' || p.id
FROM community_posts p
WHERE p.content LIKE 'Volume note #%'
  AND NOT EXISTS (SELECT 1 FROM post_comments c WHERE c.content = 'Volume comment on post ' || p.id);

INSERT INTO post_comments (post_id, user_id, content)
SELECT p.id, 623, 'Thread note ' || g || ' on the clay pin.'
FROM community_posts p
CROSS JOIN generate_series(1, 16) g
WHERE p.content LIKE 'Pinned the clay set%'
  AND NOT EXISTS (SELECT 1 FROM post_comments c WHERE c.post_id = p.id AND c.content = 'Thread note ' || g || ' on the clay pin.');

INSERT INTO post_likes (user_id, post_id)
SELECT 1000 + ((p.id + k) % 80) + 1, p.id
FROM community_posts p
CROSS JOIN generate_series(0, 2) k
ON CONFLICT DO NOTHING;

INSERT INTO community_posts (user_id, content, community_type, is_public)
SELECT 622, 'Studio log #' || g || '. Clay stills, marble stack, and the HDRI wrap for the week.', 'post', true
FROM generate_series(1, 24) g
WHERE NOT EXISTS (
  SELECT 1 FROM community_posts WHERE content = 'Studio log #' || g || '. Clay stills, marble stack, and the HDRI wrap for the week.'
);

INSERT INTO follows (follower_id, following_id)
SELECT 1000 + g, 622
FROM generate_series(1, 80) g
ON CONFLICT DO NOTHING;


INSERT INTO follows (follower_id, following_id)
SELECT 622, 1000 + g
FROM generate_series(1, 24) g
ON CONFLICT DO NOTHING;

INSERT INTO follows (follower_id, following_id)
SELECT 1000 + g, 1000 + ((g % 80) + 1)
FROM generate_series(1, 80) g
WHERE g <> ((g % 80) + 1)
ON CONFLICT DO NOTHING;
