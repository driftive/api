-- name: FindOnePendingSyncOrg :one
SELECT *
FROM git_organization_sync
WHERE next_sync < NOW() FOR
    UPDATE SKIP LOCKED
LIMIT 1;

-- name: UpdateGitOrganizationSyncStatus :one
UPDATE git_organization_sync
SET synced_at = NOW(),
    next_sync = NOW() + INTERVAL '1 day'
WHERE organization_id = $1
RETURNING *;

-- name: CreateGitOrganizationSyncIfNotExists :exec
INSERT INTO git_organization_sync (organization_id, next_sync)
VALUES ($1, NOW())
ON CONFLICT (organization_id) DO NOTHING;
