-- ClaimOnePendingSyncOrg atomically selects and claims one due row. The short next_sync
-- bump is a failure backoff: on success the loop calls UpdateGitOrganizationSyncStatus,
-- which sets the real +1 day schedule.
-- name: ClaimOnePendingSyncOrg :one
UPDATE git_organization_sync
SET next_sync = NOW() + INTERVAL '5 minutes'
WHERE organization_id = (SELECT git_organization_sync.organization_id
                         FROM git_organization_sync
                         WHERE git_organization_sync.next_sync < NOW()
                         FOR UPDATE SKIP LOCKED
                         LIMIT 1)
RETURNING *;

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
