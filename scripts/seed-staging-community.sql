-- Apply to appdb_community_staging.

INSERT INTO users (id, email, name)
VALUES
(1, 'nia@example.com', 'Nia Voss'),
(2, 'leo@example.com', 'Leo Park'),
(3, 'maya@example.com', 'Maya Chen'),
(4, 'owen@example.com', 'Owen Reid')
SELECT setval('users_id_seq', GREATEST((SELECT MAX(id) FROM users), 4));

INSERT INTO user_profiles (user_id, display_name, bio, avatar_url)
VALUES
(1, 'Nia Voss', 'Lookdev and clay. Buying once, downloading forever.', '/assets/catalog/avatar.jpg'),
(2, 'Leo Park', 'Ships small games on weekends.', '/assets/catalog/p10.jpg'),
(3, 'Maya Chen', 'Environment lighting. Collects HDRIs and marble packs.', '/assets/catalog/p08.jpg'),
(4, 'Owen Reid', 'UI kits into small tools. Quiet in the feed, loud in Figma.', '/assets/catalog/p02.jpg')
ON CONFLICT (user_id) DO UPDATE SET bio = EXCLUDED.bio, avatar_url = EXCLUDED.avatar_url;

INSERT INTO community_posts (user_id, content, community_type, is_public)
SELECT 1, 'Pinned the clay set to the top of the library. If you render stills, use the studio tier — the turntable lights actually match the HDRIs.', 'post', true
WHERE NOT EXISTS (SELECT 1 FROM community_posts WHERE content LIKE 'Pinned the clay set%');

INSERT INTO community_posts (user_id, content, community_type, is_public)
SELECT 2, 'Arcade Tile Kit + Glyph Factory is an entire jam weekend. Anyone bundling those?', 'post', true
WHERE NOT EXISTS (SELECT 1 FROM community_posts WHERE content LIKE 'Arcade Tile Kit%');

INSERT INTO community_posts (user_id, content, community_type, is_public)
SELECT 3, 'If you pin marble + clay, the stills stack is basically the Studio Kit.', 'post', true
WHERE NOT EXISTS (SELECT 1 FROM community_posts WHERE content LIKE 'If you pin marble%');

INSERT INTO follows (follower_id, following_id)
SELECT 2, 1 WHERE NOT EXISTS (SELECT 1 FROM follows WHERE follower_id=2 AND following_id=1);
INSERT INTO follows (follower_id, following_id)
SELECT 3, 1 WHERE NOT EXISTS (SELECT 1 FROM follows WHERE follower_id=3 AND following_id=1);
INSERT INTO follows (follower_id, following_id)
SELECT 4, 2 WHERE NOT EXISTS (SELECT 1 FROM follows WHERE follower_id=4 AND following_id=2);
