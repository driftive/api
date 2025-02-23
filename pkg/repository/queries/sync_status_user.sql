-- name: FindOnePendingSyncStatusUser :one
SELECT *
FROM sync_status_user
WHERE next_sync < NOW() FOR
    UPDATE SKIP LOCKED
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
