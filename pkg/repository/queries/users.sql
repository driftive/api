-- name: FindUserByID :one
SELECT *
FROM users
WHERE id = @id;

-- name: CreateUser :one
INSERT INTO users (name, email, access_token, access_token_expires_at, refresh_token, refresh_token_expires_at)
VALUES (@name, @email, @access_token, @access_token_expires_at, @refresh_token, @refresh_token_expires_at)
RETURNING *;
