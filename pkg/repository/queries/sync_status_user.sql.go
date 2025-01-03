// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: sync_status_user.sql

package queries

import (
	"context"
)

const findOnePendingSyncStatusUser = `-- name: FindOnePendingSyncStatusUser :one
SELECT id, user_id, status, synced_at
FROM sync_status_user
WHERE synced_at < NOW() - INTERVAL '1 day' FOR
UPDATE SKIP LOCKED LIMIT 1
`

func (q *Queries) FindOnePendingSyncStatusUser(ctx context.Context) (SyncStatusUser, error) {
	row := q.db.QueryRow(ctx, findOnePendingSyncStatusUser)
	var i SyncStatusUser
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Status,
		&i.SyncedAt,
	)
	return i, err
}
