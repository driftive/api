ALTER TABLE drift_analysis_run
    ADD COLUMN status           VARCHAR(20) NOT NULL DEFAULT 'COMPLETED'
        CHECK (status IN ('RUNNING', 'COMPLETED')),
    ADD COLUMN running_projects TEXT[]      NOT NULL DEFAULT '{}';

CREATE INDEX drift_analysis_run_running_updated_at_idx
    ON drift_analysis_run (updated_at)
    WHERE status = 'RUNNING';
