-- name: FindOnePendingSyncStatusUser :one
SELECT *
FROM sync_status_user
WHERE synced_at < NOW() - INTERVAL '1 day' FOR
UPDATE SKIP LOCKED LIMIT 1;
