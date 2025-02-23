// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: sync_status_user.sql

package queries

import (
	"context"
)

const createOrUpdateSyncStatusUser = `-- name: CreateOrUpdateSyncStatusUser :one
INSERT INTO sync_status_user (user_id)
VALUES ($1)
ON CONFLICT (user_id) DO NOTHING
RETURNING id, user_id, synced_at, next_sync, attempts
`

func (q *Queries) CreateOrUpdateSyncStatusUser(ctx context.Context, userID int64) (SyncStatusUser, error) {
	row := q.db.QueryRow(ctx, createOrUpdateSyncStatusUser, userID)
	var i SyncStatusUser
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.SyncedAt,
		&i.NextSync,
		&i.Attempts,
	)
	return i, err
}

const findOnePendingSyncStatusUser = `-- name: FindOnePendingSyncStatusUser :one
SELECT id, user_id, synced_at, next_sync, attempts
FROM sync_status_user
WHERE next_sync < NOW() FOR
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
		&i.NextSync,
		&i.Attempts,
	)
	return i, err
}

const findSyncStatusUserByUserID = `-- name: FindSyncStatusUserByUserID :one
SELECT id, user_id, synced_at, next_sync, attempts
FROM sync_status_user
WHERE user_id = $1
`

func (q *Queries) FindSyncStatusUserByUserID(ctx context.Context, userID int64) (SyncStatusUser, error) {
	row := q.db.QueryRow(ctx, findSyncStatusUserByUserID, userID)
	var i SyncStatusUser
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.SyncedAt,
		&i.NextSync,
		&i.Attempts,
	)
	return i, err
}

const updateSyncStatusUserLastSyncedAt = `-- name: UpdateSyncStatusUserLastSyncedAt :one
UPDATE sync_status_user
SET synced_at = NOW(),
    next_sync = NOW() + INTERVAL '30 days'
WHERE user_id = $1
RETURNING id, user_id, synced_at, next_sync, attempts
`

func (q *Queries) UpdateSyncStatusUserLastSyncedAt(ctx context.Context, userID int64) (SyncStatusUser, error) {
	row := q.db.QueryRow(ctx, updateSyncStatusUserLastSyncedAt, userID)
	var i SyncStatusUser
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.SyncedAt,
		&i.NextSync,
		&i.Attempts,
	)
	return i, err
}
