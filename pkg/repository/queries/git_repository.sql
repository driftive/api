-- name: FindGitRepositoryById :one
SELECT *
FROM git_repository
WHERE id = @id;

-- name: CreateOrUpdateRepository :one
INSERT INTO git_repository (organization_id, provider_id, name)
VALUES (@organization_id, @provider_id, @name)
ON CONFLICT (organization_id, provider_id) DO UPDATE
    SET name = @name
RETURNING *;
