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
         JOIN user_git_organization uo
              ON go.id = uo.git_organization_id
WHERE go.provider = $1
  AND uo.user_id = $2;

-- name: CreateOrUpdateGitOrganization :one
INSERT INTO git_organization (provider, provider_id, name, avatar_url)
VALUES ($1, $2, $3, $4)
ON CONFLICT (provider, provider_id) DO UPDATE
    SET provider    = $1,
        provider_id = $2,
        name        = $3,
        avatar_url  = $4
RETURNING *;

-- name: UpdateUserGitOrganizationMembership :exec
INSERT INTO user_git_organization (user_id, git_organization_id, role)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, git_organization_id) DO UPDATE
    SET role = $3;

-- name: UpdateOrgInstallationID :exec
UPDATE git_organization
SET installation_id = $2
WHERE id = $1;

-- name: IsUserMemberOfOrganization :one
SELECT EXISTS(SELECT 1
              FROM user_git_organization
              WHERE git_organization_id = $1
                AND user_id = $2);

-- name: FindGitOrganizationByRepoId :one
SELECT go.*
FROM git_organization go
         JOIN git_repository gr
              ON go.id = gr.organization_id
WHERE gr.id = $1;

-- name: IsUserMemberOfOrganizationByRepoId :one
SELECT EXISTS(SELECT 1
              FROM user_git_organization ugo
                       JOIN git_repository gr
                            ON ugo.git_organization_id = gr.organization_id
              WHERE gr.id = @repo_id
                AND ugo.user_id = @user_id);
