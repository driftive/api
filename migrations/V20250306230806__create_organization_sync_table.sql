CREATE TABLE git_organization_sync
(
    id              BIGSERIAL PRIMARY KEY,
    organization_id BIGINT      NOT NULL REFERENCES git_organization (id),
    synced_at       TIMESTAMPTZ NOT NULL DEFAULT '1970-01-01 00:00:00',
    next_sync       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_git_organization_sync_organization_id
    ON git_organization_sync (organization_id);

INSERT INTO git_organization_sync (organization_id)
SELECT id
FROM git_organization;
