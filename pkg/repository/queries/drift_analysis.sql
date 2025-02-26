-- name: CreateDriftAnalysisRun :one
INSERT INTO drift_analysis_run (uuid, repository_id, total_projects, total_projects_drifted, analysis_duration_millis)
VALUES (@uuid, @repository_id, @total_projects, @total_projects_drifted, @analysis_duration_millis)
RETURNING *;

-- name: CreateDriftAnalysisProject :one
INSERT INTO drift_analysis_project (drift_analysis_run_id, dir, type, drifted, succeeded, init_output, plan_output)
VALUES (@drift_analysis_run_id, @dir, @type, @drifted, @succeeded, @init_output, @plan_output)
RETURNING *;

-- name: FindDriftAnalysisRunsByRepositoryId :many
SELECT *
FROM drift_analysis_run
WHERE repository_id = @repository_id
ORDER BY created_at DESC
OFFSET @queryOffset LIMIT @maxResults;

-- name: FindDriftAnalysisRunByUUID :one
SELECT *
FROM drift_analysis_run
WHERE uuid = @uuid;

-- name: FindDriftAnalysisProjectsByRunId :many
SELECT *
FROM drift_analysis_project
WHERE drift_analysis_run_id = @drift_analysis_run_id;
