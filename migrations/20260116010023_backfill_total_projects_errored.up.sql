-- Backfill total_projects_errored from existing project data
UPDATE drift_analysis_run dar
SET total_projects_errored = (SELECT COUNT(*)
                              FROM drift_analysis_project dap
                              WHERE dap.drift_analysis_run_id = dar.uuid
                                AND dap.succeeded = false);
