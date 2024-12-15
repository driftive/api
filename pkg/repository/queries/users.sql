-- name: FindUserByID :one
SELECT *
FROM users
WHERE id = @id;

-- name: FindUserByProviderAndEmail :one
SELECT *
FROM users
WHERE provider = @provider
  AND email = @email;

-- name: CountUsersByProviderAndEmail :one
SELECT COUNT(*)
FROM users
WHERE provider = $1
  AND email = $2;

-- name: CreateUser :one
INSERT INTO users (provider, name, username, email, access_token, access_token_expires_at, refresh_token, refresh_token_expires_at)
VALUES (@provider, @name, @username, @email, @access_token, @access_token_expires_at, @refresh_token, @refresh_token_expires_at)
RETURNING *;
