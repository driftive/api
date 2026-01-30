ALTER TABLE drift_analysis_project
    ADD COLUMN skipped_due_to_pr BOOLEAN NOT NULL DEFAULT false;

ALTER TABLE drift_analysis_run
    ADD COLUMN total_projects_skipped INT NOT NULL DEFAULT 0;
