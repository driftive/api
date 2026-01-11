-- name: FindUserByID :one
SELECT *
FROM users
WHERE id = @id;

-- name: FindUserByProviderAndProviderId :one
SELECT *
FROM users
WHERE provider = @provider
  AND provider_id = @provider_id;

-- name: CountUsersByProviderAndProviderId :one
SELECT COUNT(*)
FROM users
WHERE provider = $1
  AND provider_id = $2;

-- name: UpsertUserOnLogin :one
INSERT INTO users (provider, provider_id, name, username, email, access_token, access_token_expires_at, refresh_token,
                   refresh_token_expires_at)
VALUES (@provider, @provider_id, @name, @username, @email, @access_token, @access_token_expires_at, @refresh_token,
        @refresh_token_expires_at) ON CONFLICT (provider, provider_id) DO
UPDATE SET
    name = @name,
    username = @username,
    email = @email,
    access_token = @access_token,
    access_token_expires_at = @access_token_expires_at,
    refresh_token = @refresh_token,
    refresh_token_expires_at = @refresh_token_expires_at,
    token_refresh_attempts = 0,
    token_refresh_disabled_at = NULL
RETURNING *;

-- name: FindExpiringTokensByProvider :many
SELECT *
FROM users
WHERE provider = @provider
  AND access_token != ''
  AND access_token_expires_at IS NOT NULL
  AND access_token_expires_at < @date
  AND refresh_token_expires_at > NOW() + INTERVAL '1 day'
  AND token_refresh_disabled_at IS NULL
LIMIT @maxResults OFFSET @queryOffset;

-- name: FindAndLockExpiringToken :one
SELECT *
FROM users
WHERE provider = @provider
  AND access_token != ''
  AND access_token_expires_at IS NOT NULL
  AND access_token_expires_at < @date
  AND refresh_token_expires_at > NOW() + INTERVAL '1 day'
  AND token_refresh_disabled_at IS NULL
LIMIT 1
FOR UPDATE SKIP LOCKED;

-- name: UpdateUserTokens :one
UPDATE users
SET access_token               = @access_token,
    access_token_expires_at    = @access_token_expires_at,
    refresh_token              = @refresh_token,
    refresh_token_expires_at   = @refresh_token_expires_at,
    token_refresh_attempts     = 0,
    token_refresh_disabled_at  = NULL
WHERE id = @id RETURNING *;

-- name: IncrementTokenRefreshAttempts :one
UPDATE users
SET token_refresh_attempts = token_refresh_attempts + 1
WHERE id = @id RETURNING *;

-- name: DisableTokenRefresh :one
UPDATE users
SET token_refresh_disabled_at = NOW()
WHERE id = @id RETURNING *;
