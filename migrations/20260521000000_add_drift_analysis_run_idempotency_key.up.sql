ALTER TABLE drift_analysis_run
    ADD COLUMN idempotency_key TEXT;

CREATE UNIQUE INDEX drift_analysis_run_repo_idem_key_uidx
    ON drift_analysis_run (repository_id, idempotency_key)
    WHERE idempotency_key IS NOT NULL;
