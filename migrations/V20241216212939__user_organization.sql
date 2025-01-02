CREATE TABLE user_organization
(
    id                  BIGSERIAL PRIMARY KEY,
    provider            VARCHAR(255) NOT NULL,
    user_id             BIGINT       NOT NULL REFERENCES users (id),
    git_organization_id BIGINT       NOT NULL REFERENCES git_organization (id)
);

CREATE UNIQUE INDEX user_organizations_provider_user_id_organization_id_idx
    ON user_organization (provider, user_id, git_organization_id);
