CREATE TABLE sync_status_organization
(
    id        BIGSERIAL PRIMARY KEY,
    provider  VARCHAR(255) NOT NULL,
    user_id   BIGINT       NOT NULL REFERENCES users (id),
    status    VARCHAR(255) NOT NULL,
    synced_at TIMESTAMPTZ  NOT NULL DEFAULT '1970-01-01 00:00:00'
);

CREATE UNIQUE INDEX sync_status_organization_provider_user_id_idx
    ON sync_status_organization (provider, user_id);
