-- Partial index to optimize FindExpiringTokensByProvider and FindAndLockExpiringToken queries
-- Filters out disabled tokens and empty access tokens
CREATE INDEX idx_users_expiring_tokens ON users (provider, access_token_expires_at)
    WHERE access_token != ''
        AND access_token_expires_at IS NOT NULL
        AND token_refresh_disabled_at IS NULL;
