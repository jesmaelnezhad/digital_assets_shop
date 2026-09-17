-- 008: key/value settings table (admin-configurable, e.g. seller wallet address)
CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- default placeholder the admin must replace with the real seller wallet
INSERT INTO settings (key, value) VALUES ('payment_address', '0xYourBSCWelcomeAddressHere')
ON CONFLICT (key) DO NOTHING;
