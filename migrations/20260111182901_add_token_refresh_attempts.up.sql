ALTER TABLE users
    ADD COLUMN token_refresh_attempts INT NOT NULL DEFAULT 0,
    ADD COLUMN token_refresh_disabled_at TIMESTAMPTZ;
