-- 009: server-side JWT revocation so logout actually logs out
CREATE TABLE IF NOT EXISTS invalidated_tokens (
    token_hash TEXT PRIMARY KEY,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_invalidated_tokens_expires ON invalidated_tokens(expires_at);
