-- name: FindGitOrganizationByID :one
SELECT *
FROM git_organization
WHERE id = $1;

-- name: FindGitOrganizationByProviderAndName :one
SELECT *
FROM git_organization
WHERE provider = $1
  AND name = $2;

-- name: FindGitOrganizationByProviderAndUserID :many
SELECT go.*
FROM git_organization go
JOIN user_organization uo
ON go.id = uo.git_organization_id
WHERE go.provider = $1
  AND uo.user_id = $2;

-- name: CreateOrUpdateGitOrganization :one
INSERT INTO git_organization (provider, provider_id, name)
VALUES ($1, $2, $3) ON CONFLICT (provider, provider_id) DO
UPDATE SET provider = $1, provider_id = $2, name = $3
    RETURNING *;
