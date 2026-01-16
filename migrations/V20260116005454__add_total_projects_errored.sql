ALTER TABLE drift_analysis_run
    ADD COLUMN total_projects_errored INT NOT NULL DEFAULT 0;
