-- Community users must share identity-service IDs so JWT user_id matches the feed.

INSERT INTO users (id, email, name) VALUES
(622, 'nia@example.com', 'Nia Voss'),
(623, 'leo@example.com', 'Leo Park'),
(624, 'maya@example.com', 'Maya Chen'),
(625, 'owen@example.com', 'Owen Reid')
ON CONFLICT (id) DO UPDATE SET email = EXCLUDED.email, name = EXCLUDED.name;

SELECT setval('users_id_seq', GREATEST((SELECT MAX(id) FROM users), 625));

INSERT INTO user_profiles (user_id, display_name, bio, avatar_url) VALUES
(622, 'Nia Voss', 'Lookdev and clay. Buying once, downloading forever.', '/assets/catalog/avatar.jpg'),
(623, 'Leo Park', 'Ships small games on weekends.', '/assets/catalog/p10.jpg'),
(624, 'Maya Chen', 'Environment lighting. Collects HDRIs and marble packs.', '/assets/catalog/p08.jpg'),
(625, 'Owen Reid', 'UI kits into small tools. Quiet in the feed, loud in Figma.', '/assets/catalog/p02.jpg')
ON CONFLICT (user_id) DO UPDATE SET
  display_name = EXCLUDED.display_name,
  bio = EXCLUDED.bio,
  avatar_url = EXCLUDED.avatar_url;

DELETE FROM community_posts WHERE content = 'Test post content';

INSERT INTO community_posts (user_id, content, community_type, is_public)
SELECT 622, 'Pinned the clay set to the top of the library. If you render stills, use the studio tier — the turntable lights actually match the HDRIs.', 'post', true
WHERE NOT EXISTS (SELECT 1 FROM community_posts WHERE content LIKE 'Pinned the clay set%');

INSERT INTO community_posts (user_id, content, community_type, is_public)
SELECT 623, 'Arcade Tile Kit + Glyph Factory is an entire jam weekend. Anyone bundling those?', 'post', true
WHERE NOT EXISTS (SELECT 1 FROM community_posts WHERE content LIKE 'Arcade Tile Kit%');

INSERT INTO community_posts (user_id, content, community_type, is_public)
SELECT 624, 'If you pin marble + clay, the stills stack is basically the Studio Kit.', 'post', true
WHERE NOT EXISTS (SELECT 1 FROM community_posts WHERE content LIKE 'If you pin marble%');

INSERT INTO follows (follower_id, following_id)
SELECT 623, 622 WHERE NOT EXISTS (SELECT 1 FROM follows WHERE follower_id=623 AND following_id=622);
INSERT INTO follows (follower_id, following_id)
SELECT 624, 622 WHERE NOT EXISTS (SELECT 1 FROM follows WHERE follower_id=624 AND following_id=622);
INSERT INTO follows (follower_id, following_id)
SELECT 625, 623 WHERE NOT EXISTS (SELECT 1 FROM follows WHERE follower_id=625 AND following_id=623);
