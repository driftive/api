-- name: FindOnePendingSyncStatusUser :one
SELECT sync_status_user.*
FROM sync_status_user
         INNER JOIN users ON users.id = sync_status_user.user_id
WHERE sync_status_user.next_sync < NOW()
  AND users.token_refresh_disabled_at IS NULL
FOR UPDATE OF sync_status_user SKIP LOCKED
LIMIT 1;


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
