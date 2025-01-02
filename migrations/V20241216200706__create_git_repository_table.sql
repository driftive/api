CREATE TABLE git_repository
(
    id              BIGSERIAL PRIMARY KEY,
    organization_id BIGINT       NOT NULL REFERENCES git_organization (id),
    name            VARCHAR(255) NOT NULL,
    analysis_token  VARCHAR(255) NOT NULL
);
