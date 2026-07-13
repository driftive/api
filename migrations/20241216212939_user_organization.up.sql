CREATE TABLE user_git_organization
(
    id                  BIGSERIAL PRIMARY KEY,
    user_id             BIGINT       NOT NULL REFERENCES users (id),
    git_organization_id BIGINT       NOT NULL REFERENCES git_organization (id),
    role                VARCHAR(255) NOT NULL
);

CREATE UNIQUE INDEX user_git_organizations_provider_user_id_organization_id_idx
    ON user_git_organization (user_id, git_organization_id);
