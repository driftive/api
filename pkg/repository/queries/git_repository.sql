-- name: FindGitRepositoryById :one
SELECT *
FROM git_repository
WHERE id = @id;

-- name: CreateOrUpdateRepository :one
INSERT INTO git_repository (organization_id, provider_id, name, is_private)
VALUES (@organization_id, @provider_id, @name, @is_private)
ON CONFLICT (organization_id, provider_id) DO UPDATE
    SET name       = @name,
        is_private = @is_private
RETURNING *;

-- name: FindGitRepositoriesByOrgId :many
SELECT *
FROM git_repository
WHERE organization_id = @organization_id
ORDER BY (analysis_token IS NOT NULL) DESC, name ASC;

-- name: FindGitRepositoryByOrgIdAndName :one
SELECT *
FROM git_repository
WHERE organization_id = @organization_id
  AND name = @name;

-- name: UpdateRepositoryToken :one
UPDATE git_repository
SET analysis_token = @token
WHERE id = @id
RETURNING analysis_token;

-- name: FindGitRepositoryByToken :one
SELECT *
FROM git_repository
WHERE analysis_token = $1
  AND analysis_token IS NOT NULL
  AND analysis_token != '';
