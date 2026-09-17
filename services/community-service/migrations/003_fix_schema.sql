-- Fix community DB tables for user_profiles (missing wallet_address and bio)
ALTER TABLE user_profiles ADD COLUMN IF NOT EXISTS wallet_address VARCHAR(255) DEFAULT '';
ALTER TABLE user_profiles ADD COLUMN IF NOT EXISTS bio TEXT DEFAULT '';

-- Fix admin DB tables
-- Products: add pinned column
ALTER TABLE products ADD COLUMN IF NOT EXISTS pinned BOOLEAN DEFAULT false;

-- Settings: add unique constraint on key if not exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'settings_key_key'
    ) THEN
        ALTER TABLE settings ADD CONSTRAINT settings_key_key UNIQUE (key);
    END IF;
END
$$;
