-- Apply to appdb_identity_staging.

INSERT INTO users (email, password_hash, name) VALUES
('nia@example.com', '$2b$10$blu/StVHqw2O.hZO7HrYEuETFYgw8bICkm1QMChmFVU.i.bT94AZa', 'Nia Voss'),
('leo@example.com', '$2b$10$ZCQ1OUAvU26XJf72mUualu3ldk2LuwVLKD3A67IEqCkQVEeJOcvle', 'Leo Park'),
('maya@example.com', '$2b$10$Yfoej/JpgeIxK4iZe92bqu0mlKPkqmQWiHzhvLEgImehCsrxJJorS', 'Maya Chen'),
('owen@example.com', '$2b$10$x8cLjAINAeMbQdQXVkTvguiLHB1OR96pDhALZgvswrKeTBS122gje', 'Owen Reid')
ON CONFLICT (email) DO UPDATE SET name = EXCLUDED.name, password_hash = EXCLUDED.password_hash;

INSERT INTO user_profiles (user_id, display_name, bio, wallet_address, avatar_url)
SELECT id, name,
  CASE email
    WHEN 'nia@example.com' THEN 'Lookdev and clay. Buying once, downloading forever.'
    WHEN 'leo@example.com' THEN 'Ships small games on weekends.'
    WHEN 'maya@example.com' THEN 'Environment lighting. Collects HDRIs and marble packs.'
    ELSE 'UI kits into small tools. Quiet in the feed, loud in Figma.'
  END,
  CASE email
    WHEN 'nia@example.com' THEN '0xNIA00a4f'
    WHEN 'maya@example.com' THEN '0xMAYAc2e'
    ELSE ''
  END,
  CASE email
    WHEN 'nia@example.com' THEN '/assets/catalog/avatar.jpg'
    WHEN 'leo@example.com' THEN '/assets/catalog/p10.jpg'
    WHEN 'maya@example.com' THEN '/assets/catalog/p08.jpg'
    ELSE '/assets/catalog/p02.jpg'
  END
FROM users WHERE email IN ('nia@example.com','leo@example.com','maya@example.com','owen@example.com')
ON CONFLICT (user_id) DO UPDATE SET bio = EXCLUDED.bio, avatar_url = EXCLUDED.avatar_url, wallet_address = EXCLUDED.wallet_address;

INSERT INTO referral_links (user_id, code, is_active)
SELECT id, CASE email
  WHEN 'nia@example.com' THEN 'NIA-STUDIO'
  WHEN 'leo@example.com' THEN 'LEO-BUSY'
  WHEN 'maya@example.com' THEN 'MAYA-LIGHT'
  ELSE 'OWEN-GRID'
END, true
FROM users WHERE email IN ('nia@example.com','leo@example.com','maya@example.com','owen@example.com')
ON CONFLICT (code) DO NOTHING;
