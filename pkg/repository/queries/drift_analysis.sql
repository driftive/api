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

-- name: GetRepositoryRunStats :one
SELECT
    COUNT(*) AS total_runs,
    COUNT(*) FILTER (WHERE total_projects_drifted > 0) AS runs_with_drift,
    MAX(created_at) AS last_run_at
FROM drift_analysis_run
WHERE repository_id = @repository_id;

-- name: GetLatestRunForRepository :one
SELECT *
FROM drift_analysis_run
WHERE repository_id = @repository_id
ORDER BY created_at DESC
LIMIT 1;

-- name: GetDriftRateOverTime :many
-- Returns daily drift rate data for the specified time range
SELECT
    DATE(created_at) AS date,
    COUNT(*)::BIGINT AS total_runs,
    COUNT(*) FILTER (WHERE total_projects_drifted > 0)::BIGINT AS runs_with_drift
FROM drift_analysis_run
WHERE repository_id = @repository_id
  AND created_at >= NOW() - (sqlc.arg(days_back)::INTEGER || ' days')::INTERVAL
GROUP BY DATE(created_at)
ORDER BY DATE(created_at) ASC;

-- name: GetMostFrequentlyDriftedProjects :many
-- Returns projects ranked by how often they drift (top N)
SELECT
    dap.dir,
    dap.type,
    COUNT(*) FILTER (WHERE dap.drifted = true)::BIGINT AS drift_count,
    COUNT(*)::BIGINT AS total_appearances
FROM drift_analysis_project dap
JOIN drift_analysis_run dar ON dap.drift_analysis_run_id = dar.uuid
WHERE dar.repository_id = @repository_id
  AND dar.created_at >= NOW() - (sqlc.arg(days_back)::INTEGER || ' days')::INTERVAL
GROUP BY dap.dir, dap.type
HAVING COUNT(*) FILTER (WHERE dap.drifted = true) > 0
ORDER BY drift_count DESC
LIMIT sqlc.arg(max_results);

-- name: GetDriftFreeStreak :one
-- Returns the current consecutive run count without drift
WITH ranked_runs AS (
    SELECT
        uuid,
        total_projects_drifted,
        created_at,
        ROW_NUMBER() OVER (ORDER BY created_at DESC) AS rn
    FROM drift_analysis_run
    WHERE repository_id = @repository_id
),
first_drift AS (
    SELECT MIN(rn) AS break_point
    FROM ranked_runs
    WHERE total_projects_drifted > 0
)
SELECT
    COALESCE(
        (SELECT break_point - 1 FROM first_drift WHERE break_point IS NOT NULL),
        (SELECT COUNT(*) FROM ranked_runs)
    )::BIGINT AS streak_count,
    (SELECT created_at FROM ranked_runs WHERE rn = 1) AS last_run_at;

-- name: GetMeanTimeToResolution :many
-- Returns drift resolution times by tracking when a drifted project becomes non-drifted
WITH project_states AS (
    SELECT
        dap.dir,
        dar.created_at,
        dap.drifted,
        LAG(dap.drifted) OVER (PARTITION BY dap.dir ORDER BY dar.created_at) AS prev_drifted,
        LAG(dar.created_at) OVER (PARTITION BY dap.dir ORDER BY dar.created_at) AS prev_created_at
    FROM drift_analysis_project dap
    JOIN drift_analysis_run dar ON dap.drift_analysis_run_id = dar.uuid
    WHERE dar.repository_id = @repository_id
      AND dar.created_at >= NOW() - (sqlc.arg(days_back)::INTEGER || ' days')::INTERVAL
),
resolutions AS (
    SELECT
        dir,
        created_at AS resolved_at,
        prev_created_at AS drifted_at,
        EXTRACT(EPOCH FROM (created_at - prev_created_at)) / 3600 AS hours_to_resolve
    FROM project_states
    WHERE prev_drifted = true AND drifted = false
)
SELECT
    DATE(resolved_at) AS date,
    COUNT(*)::BIGINT AS resolutions_count,
    AVG(hours_to_resolve) AS avg_hours_to_resolve
FROM resolutions
GROUP BY DATE(resolved_at)
ORDER BY DATE(resolved_at) ASC;
