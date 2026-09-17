-- Add missing tables to community database
-- The community service needs to reference users and user_profiles
-- Since each service has its own DB, we add minimal copies

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL DEFAULT '',
    name VARCHAR(255) NOT NULL DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_profiles (
    user_id INTEGER PRIMARY KEY,
    display_name VARCHAR(255) DEFAULT '',
    avatar_url TEXT DEFAULT '',
    bio TEXT DEFAULT '',
    location VARCHAR(255) DEFAULT '',
    website VARCHAR(500) DEFAULT '',
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
