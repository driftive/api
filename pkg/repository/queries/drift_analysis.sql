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
WHERE drift_analysis_run_id = @drift_analysis_run_id
ORDER BY
    CASE
        WHEN succeeded = false THEN 0  -- Errored first
        WHEN drifted = true THEN 1     -- Drifted second
        ELSE 2                         -- OK last
    END,
    dir ASC;

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
-- Returns drift resolution times by tracking from when drift STARTED to when it was resolved
WITH project_states AS (
    SELECT
        dap.dir,
        dar.created_at,
        dap.drifted,
        LAG(dap.drifted) OVER (PARTITION BY dap.dir ORDER BY dar.created_at) AS prev_drifted
    FROM drift_analysis_project dap
    JOIN drift_analysis_run dar ON dap.drift_analysis_run_id = dar.uuid
    WHERE dar.repository_id = @repository_id
),
-- Mark transitions: drift_start when going from not-drifted to drifted, drift_end when going from drifted to not-drifted
transitions AS (
    SELECT
        dir,
        created_at,
        drifted,
        CASE WHEN (prev_drifted IS NULL OR prev_drifted = false) AND drifted = true THEN created_at END AS drift_start,
        CASE WHEN prev_drifted = true AND drifted = false THEN created_at END AS drift_end
    FROM project_states
),
-- Get drift start times
drift_starts AS (
    SELECT dir, drift_start, ROW_NUMBER() OVER (PARTITION BY dir ORDER BY drift_start) AS start_num
    FROM transitions
    WHERE drift_start IS NOT NULL
),
-- Get drift end times
drift_ends AS (
    SELECT dir, drift_end, ROW_NUMBER() OVER (PARTITION BY dir ORDER BY drift_end) AS end_num
    FROM transitions
    WHERE drift_end IS NOT NULL
),
-- Match each resolution with its corresponding drift start
matched_resolutions AS (
    SELECT
        e.dir,
        s.drift_start,
        e.drift_end AS resolved_at,
        EXTRACT(EPOCH FROM (e.drift_end - s.drift_start)) / 3600 AS hours_to_resolve
    FROM drift_ends e
    JOIN drift_starts s ON e.dir = s.dir AND e.end_num = s.start_num
    WHERE e.drift_end >= NOW() - (sqlc.arg(days_back)::INTEGER || ' days')::INTERVAL
)
SELECT
    DATE(resolved_at) AS date,
    COUNT(*)::BIGINT AS resolutions_count,
    AVG(hours_to_resolve) AS avg_hours_to_resolve
FROM matched_resolutions
GROUP BY DATE(resolved_at)
ORDER BY DATE(resolved_at) ASC;

-- name: DeleteOldestRunsExceedingLimit :exec
-- Deletes the oldest runs for a repository, keeping only the most recent N runs
DELETE FROM drift_analysis_run dar
WHERE dar.uuid IN (
    SELECT r.uuid FROM drift_analysis_run r
    WHERE r.repository_id = @repository_id
    ORDER BY r.created_at DESC
    OFFSET @max_runs_to_keep
);
