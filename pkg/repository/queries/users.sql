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
