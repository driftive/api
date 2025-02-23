CREATE TABLE git_organization
(
    id              BIGSERIAL PRIMARY KEY,
    provider        VARCHAR(255) NOT NULL,
    provider_id     VARCHAR(255) NOT NULL CHECK (provider_id != '0'),
    name            VARCHAR(255) NOT NULL,
    avatar_url      VARCHAR,
    installation_id BIGINT
);

CREATE UNIQUE INDEX git_organizations_provider_provider_id_idx
    ON git_organization (provider, provider_id);
