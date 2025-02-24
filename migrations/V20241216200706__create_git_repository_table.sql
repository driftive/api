CREATE TABLE git_repository
(
    id              BIGSERIAL PRIMARY KEY,
    organization_id BIGINT       NOT NULL REFERENCES git_organization (id),
    provider_id     VARCHAR      NOT NULL,
    name            VARCHAR(255) NOT NULL,
    is_private      BOOLEAN      NOT NULL,
    analysis_token  VARCHAR(255)
);

CREATE UNIQUE INDEX git_repositories_provider_id_organization_id_idx
    ON git_repository (organization_id, provider_id);

CREATE UNIQUE INDEX git_repositories_analysis_token_idx
    ON git_repository (analysis_token);

