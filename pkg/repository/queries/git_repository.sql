-- name: FindGitRepositoryById :one
SELECT *
FROM git_repository
WHERE id = @id;
