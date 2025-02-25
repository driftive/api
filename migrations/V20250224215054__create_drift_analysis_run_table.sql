CREATE TABLE drift_analysis_run
(
    uuid                     UUID PRIMARY KEY,
    repository_id            BIGINT      NOT NULL REFERENCES git_repository (id),
    total_projects           INT         NOT NULL,
    total_projects_drifted   INT         NOT NULL,
    analysis_duration_millis BIGINT      NOT NULL,
    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW()
)
