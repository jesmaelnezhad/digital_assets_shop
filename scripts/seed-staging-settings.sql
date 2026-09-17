-- Apply to appdb_payment_staging and/or appdb_admin_staging settings.

INSERT INTO settings (key, value, updated_at) VALUES
('payment_address', '0xPAWRADISE_WALLET_BSC', NOW()),
('site_name', 'Pawradise', NOW()),
('site_title', 'Pawradise — digital assets', NOW()),
('site_description', 'Buy once, download forever. Clay characters, UI kits, textures, and scenes.', NOW()),
('download_policy', 'Buy once, download forever. Unlimited re-downloads on paid orders.', NOW()),
('default_currency', 'USD', NOW()),
('robots_index', 'true', NOW()),
('referral_commission_percent', '5.0', NOW())
ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW();
