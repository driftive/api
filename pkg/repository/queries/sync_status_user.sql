-- name: FindOnePendingSyncStatusUser :one
SELECT *
FROM sync_status_user
WHERE synced_at < NOW() - INTERVAL '1 day' FOR
UPDATE SKIP LOCKED LIMIT 1;


-- name: CreateSyncStatusUser :one
INSERT INTO sync_status_user (user_id, synced_at)
VALUES ($1, $2) RETURNING *;
