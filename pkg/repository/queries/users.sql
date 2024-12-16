-- name: FindUserByID :one
SELECT *
FROM users
WHERE id = @id;

-- name: FindUserByProviderAndProviderId :one
SELECT *
FROM users
WHERE provider = @provider
  AND provider_id = @provider_id;

-- name: CountUsersByProviderAndEmail :one
SELECT COUNT(*)
FROM users
WHERE provider = $1
  AND email = $2;

-- name: CreateUser :one
INSERT INTO users (provider, provider_id, name, username, email, access_token, access_token_expires_at, refresh_token, refresh_token_expires_at)
VALUES (@provider, @provider_id, @name, @username, @email, @access_token, @access_token_expires_at, @refresh_token, @refresh_token_expires_at)
RETURNING *;

-- name: FindExpiringTokensByProvider :many
SELECT *
FROM users
WHERE provider = @provider
  AND access_token_expires_at IS NOT NULL
  AND access_token_expires_at < @date LIMIT @maxResults
OFFSET @queryOffset;

-- name: UpdateUserTokens :one
UPDATE users
SET access_token             = @access_token,
    access_token_expires_at  = @access_token_expires_at,
    refresh_token            = @refresh_token,
    refresh_token_expires_at = @refresh_token_expires_at
WHERE id = @id RETURNING *;
