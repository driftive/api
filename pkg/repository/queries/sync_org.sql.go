// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: sync_org.sql

package queries

import (
	"context"
)

const findOnePendingSyncOrg = `-- name: FindOnePendingSyncOrg :one
SELECT id, organization_id, synced_at, next_sync
FROM git_organization_sync
WHERE next_sync < NOW() FOR
    UPDATE SKIP LOCKED
LIMIT 1
`

func (q *Queries) FindOnePendingSyncOrg(ctx context.Context) (GitOrganizationSync, error) {
	row := q.db.QueryRow(ctx, findOnePendingSyncOrg)
	var i GitOrganizationSync
	err := row.Scan(
		&i.ID,
		&i.OrganizationID,
		&i.SyncedAt,
		&i.NextSync,
	)
	return i, err
}

const updateGitOrganizationSyncStatus = `-- name: UpdateGitOrganizationSyncStatus :one
UPDATE git_organization_sync
SET synced_at = NOW(),
    next_sync = NOW() + INTERVAL '1 day'
WHERE organization_id = $1
RETURNING id, organization_id, synced_at, next_sync
`

func (q *Queries) UpdateGitOrganizationSyncStatus(ctx context.Context, organizationID int64) (GitOrganizationSync, error) {
	row := q.db.QueryRow(ctx, updateGitOrganizationSyncStatus, organizationID)
	var i GitOrganizationSync
	err := row.Scan(
		&i.ID,
		&i.OrganizationID,
		&i.SyncedAt,
		&i.NextSync,
	)
	return i, err
}
