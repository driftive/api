-- ClaimOnePendingSyncStatusUser atomically selects and claims one due row. The short
-- next_sync bump is a failure backoff: on success SyncUserResources calls
-- UpdateSyncStatusUserLastSyncedAt, which sets the real +30 day schedule.
-- name: ClaimOnePendingSyncStatusUser :one
UPDATE sync_status_user
SET next_sync = NOW() + INTERVAL '5 minutes'
WHERE user_id = (SELECT sync_status_user.user_id
                 FROM sync_status_user
                          INNER JOIN users ON users.id = sync_status_user.user_id
                 WHERE sync_status_user.next_sync < NOW()
                   AND users.token_refresh_disabled_at IS NULL
                 FOR UPDATE OF sync_status_user SKIP LOCKED
                 LIMIT 1)
RETURNING *;

-- name: CreateOrUpdateSyncStatusUser :one
INSERT INTO sync_status_user (user_id)
VALUES ($1)
ON CONFLICT (user_id) DO NOTHING
RETURNING *;

-- name: UpdateSyncStatusUserLastSyncedAt :one
UPDATE sync_status_user
SET synced_at = NOW(),
    next_sync = NOW() + INTERVAL '30 days'
WHERE user_id = $1
RETURNING *;

-- name: FindSyncStatusUserByUserID :one
SELECT *
FROM sync_status_user
WHERE user_id = $1;
