CREATE TABLE users
(
    id                       BIGSERIAL PRIMARY KEY,
    name                     VARCHAR      NOT NULL,
    username                 VARCHAR(255) NOT NULL,
    email                    VARCHAR(255) NOT NULL,
    access_token             VARCHAR(255) NOT NULL,
    access_token_expires_at  TIMESTAMPTZ,
    refresh_token            VARCHAR(255) NOT NULL,
    refresh_token_expires_at TIMESTAMPTZ
);
