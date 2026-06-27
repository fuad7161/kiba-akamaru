-- One active refresh token per user (upsert strategy).
-- Expires column lets us enforce TTL purely in SQL.
CREATE TABLE refresh_tokens (
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token       TEXT        NOT NULL,
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (user_id)
);

-- Index on token so we can look up by token value during refresh/logout.
CREATE INDEX idx_refresh_tokens_token ON refresh_tokens(token);
