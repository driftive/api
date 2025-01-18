-- name: FindOnePendingSyncStatusUser :one
SELECT *
FROM sync_status_user
WHERE synced_at < NOW() - INTERVAL '1 day' FOR
    UPDATE SKIP LOCKED
LIMIT 1;


-- name: CreateSyncStatusUser :one
INSERT INTO sync_status_user (user_id)
VALUES ($1)
RETURNING *;

-- name: UpdateSyncStatusUserLastSyncedAt :one
UPDATE sync_status_user
SET synced_at = NOW()
WHERE user_id = $1
RETURNING *;
