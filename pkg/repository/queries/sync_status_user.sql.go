// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: sync_status_user.sql

package queries

import (
	"context"
)

const createSyncStatusUser = `-- name: CreateSyncStatusUser :one
INSERT INTO sync_status_user (user_id)
VALUES ($1)
RETURNING id, user_id, synced_at, attempts
`

func (q *Queries) CreateSyncStatusUser(ctx context.Context, userID int64) (SyncStatusUser, error) {
	row := q.db.QueryRow(ctx, createSyncStatusUser, userID)
	var i SyncStatusUser
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.SyncedAt,
		&i.Attempts,
	)
	return i, err
}

const findOnePendingSyncStatusUser = `-- name: FindOnePendingSyncStatusUser :one
SELECT id, user_id, synced_at, attempts
FROM sync_status_user
WHERE synced_at < NOW() - INTERVAL '1 day' FOR
    UPDATE SKIP LOCKED
LIMIT 1
`

func (q *Queries) FindOnePendingSyncStatusUser(ctx context.Context) (SyncStatusUser, error) {
	row := q.db.QueryRow(ctx, findOnePendingSyncStatusUser)
	var i SyncStatusUser
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.SyncedAt,
		&i.Attempts,
	)
	return i, err
}

const updateSyncStatusUserLastSyncedAt = `-- name: UpdateSyncStatusUserLastSyncedAt :one
UPDATE sync_status_user
SET synced_at = NOW()
WHERE user_id = $1
RETURNING id, user_id, synced_at, attempts
`

func (q *Queries) UpdateSyncStatusUserLastSyncedAt(ctx context.Context, userID int64) (SyncStatusUser, error) {
	row := q.db.QueryRow(ctx, updateSyncStatusUserLastSyncedAt, userID)
	var i SyncStatusUser
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.SyncedAt,
		&i.Attempts,
	)
	return i, err
}
