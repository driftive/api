CREATE TABLE users
(
    id                       BIGSERIAL PRIMARY KEY,
    provider                 VARCHAR(255) NOT NULL,
    provider_id              VARCHAR      NOT NULL,
    name                     VARCHAR      NOT NULL,
    username                 VARCHAR(255) NOT NULL,
    email                    VARCHAR(255) NOT NULL,
    access_token             VARCHAR(255) NOT NULL,
    access_token_expires_at  TIMESTAMPTZ,
    refresh_token            VARCHAR(255) NOT NULL,
    refresh_token_expires_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_users_provider_provider_id ON users (provider, provider_id);
