// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: users.sql

package queries

import (
	"context"
	"time"
)

const countUsersByProviderAndEmail = `-- name: CountUsersByProviderAndEmail :one
SELECT COUNT(*)
FROM users
WHERE provider = $1
  AND email = $2
`

type CountUsersByProviderAndEmailParams struct {
	Provider string
	Email    string
}

func (q *Queries) CountUsersByProviderAndEmail(ctx context.Context, arg CountUsersByProviderAndEmailParams) (int64, error) {
	row := q.db.QueryRow(ctx, countUsersByProviderAndEmail, arg.Provider, arg.Email)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const createUser = `-- name: CreateUser :one
INSERT INTO users (provider, provider_id, name, username, email, access_token, access_token_expires_at, refresh_token, refresh_token_expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, provider, provider_id, name, username, email, access_token, access_token_expires_at, refresh_token, refresh_token_expires_at
`

type CreateUserParams struct {
	Provider              string
	ProviderID            string
	Name                  string
	Username              string
	Email                 string
	AccessToken           string
	AccessTokenExpiresAt  *time.Time
	RefreshToken          string
	RefreshTokenExpiresAt *time.Time
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRow(ctx, createUser,
		arg.Provider,
		arg.ProviderID,
		arg.Name,
		arg.Username,
		arg.Email,
		arg.AccessToken,
		arg.AccessTokenExpiresAt,
		arg.RefreshToken,
		arg.RefreshTokenExpiresAt,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Provider,
		&i.ProviderID,
		&i.Name,
		&i.Username,
		&i.Email,
		&i.AccessToken,
		&i.AccessTokenExpiresAt,
		&i.RefreshToken,
		&i.RefreshTokenExpiresAt,
	)
	return i, err
}

const findUserByID = `-- name: FindUserByID :one
SELECT id, provider, provider_id, name, username, email, access_token, access_token_expires_at, refresh_token, refresh_token_expires_at
FROM users
WHERE id = $1
`

func (q *Queries) FindUserByID(ctx context.Context, id int64) (User, error) {
	row := q.db.QueryRow(ctx, findUserByID, id)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Provider,
		&i.ProviderID,
		&i.Name,
		&i.Username,
		&i.Email,
		&i.AccessToken,
		&i.AccessTokenExpiresAt,
		&i.RefreshToken,
		&i.RefreshTokenExpiresAt,
	)
	return i, err
}

const findUserByProviderAndProviderId = `-- name: FindUserByProviderAndProviderId :one
SELECT id, provider, provider_id, name, username, email, access_token, access_token_expires_at, refresh_token, refresh_token_expires_at
FROM users
WHERE provider = $1
  AND provider_id = $2
`

type FindUserByProviderAndProviderIdParams struct {
	Provider   string
	ProviderID string
}

func (q *Queries) FindUserByProviderAndProviderId(ctx context.Context, arg FindUserByProviderAndProviderIdParams) (User, error) {
	row := q.db.QueryRow(ctx, findUserByProviderAndProviderId, arg.Provider, arg.ProviderID)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Provider,
		&i.ProviderID,
		&i.Name,
		&i.Username,
		&i.Email,
		&i.AccessToken,
		&i.AccessTokenExpiresAt,
		&i.RefreshToken,
		&i.RefreshTokenExpiresAt,
	)
	return i, err
}
