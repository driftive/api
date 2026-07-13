CREATE TABLE drift_analysis_project
(
    id                    BIGSERIAL PRIMARY KEY,
    drift_analysis_run_id UUID          NOT NULL REFERENCES drift_analysis_run (uuid),
    dir                   VARCHAR(1500) NOT NULL,
    type                  VARCHAR       NOT NULL CHECK ( type IN ('TERRAFORM', 'TOFU', 'TERRAGRUNT') ),
    drifted               BOOLEAN       NOT NULL,
    succeeded             BOOLEAN       NOT NULL,
    init_output           TEXT,
    plan_output           TEXT
);

CREATE UNIQUE INDEX drift_analysis_project_drift_analysis_run_id_dir_idx
    ON drift_analysis_project (drift_analysis_run_id, dir);
