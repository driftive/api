CREATE INDEX drift_analysis_run_repository_id_created_at_idx
    ON drift_analysis_run (repository_id, created_at DESC);
