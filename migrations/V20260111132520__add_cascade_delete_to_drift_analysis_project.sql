-- Drop the existing FK constraint and recreate with ON DELETE CASCADE
ALTER TABLE drift_analysis_project
    DROP CONSTRAINT drift_analysis_project_drift_analysis_run_id_fkey;

ALTER TABLE drift_analysis_project
    ADD CONSTRAINT drift_analysis_project_drift_analysis_run_id_fkey
        FOREIGN KEY (drift_analysis_run_id)
            REFERENCES drift_analysis_run (uuid)
            ON DELETE CASCADE;
