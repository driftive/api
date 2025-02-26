// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: drift_analysis.sql

package queries

import (
	"context"

	"github.com/google/uuid"
)

const createDriftAnalysisProject = `-- name: CreateDriftAnalysisProject :one
INSERT INTO drift_analysis_project (drift_analysis_run_id, dir, type, drifted, succeeded, init_output, plan_output)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, drift_analysis_run_id, dir, type, drifted, succeeded, init_output, plan_output
`

type CreateDriftAnalysisProjectParams struct {
	DriftAnalysisRunID uuid.UUID
	Dir                string
	Type               string
	Drifted            bool
	Succeeded          bool
	InitOutput         *string
	PlanOutput         *string
}

func (q *Queries) CreateDriftAnalysisProject(ctx context.Context, arg CreateDriftAnalysisProjectParams) (DriftAnalysisProject, error) {
	row := q.db.QueryRow(ctx, createDriftAnalysisProject,
		arg.DriftAnalysisRunID,
		arg.Dir,
		arg.Type,
		arg.Drifted,
		arg.Succeeded,
		arg.InitOutput,
		arg.PlanOutput,
	)
	var i DriftAnalysisProject
	err := row.Scan(
		&i.ID,
		&i.DriftAnalysisRunID,
		&i.Dir,
		&i.Type,
		&i.Drifted,
		&i.Succeeded,
		&i.InitOutput,
		&i.PlanOutput,
	)
	return i, err
}

const createDriftAnalysisRun = `-- name: CreateDriftAnalysisRun :one
INSERT INTO drift_analysis_run (uuid, repository_id, total_projects, total_projects_drifted, analysis_duration_millis)
VALUES ($1, $2, $3, $4, $5)
RETURNING uuid, repository_id, total_projects, total_projects_drifted, analysis_duration_millis, created_at, updated_at
`

type CreateDriftAnalysisRunParams struct {
	Uuid                   uuid.UUID
	RepositoryID           int64
	TotalProjects          int32
	TotalProjectsDrifted   int32
	AnalysisDurationMillis int64
}

func (q *Queries) CreateDriftAnalysisRun(ctx context.Context, arg CreateDriftAnalysisRunParams) (DriftAnalysisRun, error) {
	row := q.db.QueryRow(ctx, createDriftAnalysisRun,
		arg.Uuid,
		arg.RepositoryID,
		arg.TotalProjects,
		arg.TotalProjectsDrifted,
		arg.AnalysisDurationMillis,
	)
	var i DriftAnalysisRun
	err := row.Scan(
		&i.Uuid,
		&i.RepositoryID,
		&i.TotalProjects,
		&i.TotalProjectsDrifted,
		&i.AnalysisDurationMillis,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const findDriftAnalysisProjectsByRunId = `-- name: FindDriftAnalysisProjectsByRunId :many
SELECT id, drift_analysis_run_id, dir, type, drifted, succeeded, init_output, plan_output
FROM drift_analysis_project
WHERE drift_analysis_run_id = $1
`

func (q *Queries) FindDriftAnalysisProjectsByRunId(ctx context.Context, driftAnalysisRunID uuid.UUID) ([]DriftAnalysisProject, error) {
	rows, err := q.db.Query(ctx, findDriftAnalysisProjectsByRunId, driftAnalysisRunID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []DriftAnalysisProject
	for rows.Next() {
		var i DriftAnalysisProject
		if err := rows.Scan(
			&i.ID,
			&i.DriftAnalysisRunID,
			&i.Dir,
			&i.Type,
			&i.Drifted,
			&i.Succeeded,
			&i.InitOutput,
			&i.PlanOutput,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const findDriftAnalysisRunByUUID = `-- name: FindDriftAnalysisRunByUUID :one
SELECT uuid, repository_id, total_projects, total_projects_drifted, analysis_duration_millis, created_at, updated_at
FROM drift_analysis_run
WHERE uuid = $1
`

func (q *Queries) FindDriftAnalysisRunByUUID(ctx context.Context, argUuid uuid.UUID) (DriftAnalysisRun, error) {
	row := q.db.QueryRow(ctx, findDriftAnalysisRunByUUID, argUuid)
	var i DriftAnalysisRun
	err := row.Scan(
		&i.Uuid,
		&i.RepositoryID,
		&i.TotalProjects,
		&i.TotalProjectsDrifted,
		&i.AnalysisDurationMillis,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const findDriftAnalysisRunsByRepositoryId = `-- name: FindDriftAnalysisRunsByRepositoryId :many
SELECT uuid, repository_id, total_projects, total_projects_drifted, analysis_duration_millis, created_at, updated_at
FROM drift_analysis_run
WHERE repository_id = $1
ORDER BY created_at DESC
OFFSET $2 LIMIT $3
`

type FindDriftAnalysisRunsByRepositoryIdParams struct {
	RepositoryID int64
	Queryoffset  int32
	Maxresults   int32
}

func (q *Queries) FindDriftAnalysisRunsByRepositoryId(ctx context.Context, arg FindDriftAnalysisRunsByRepositoryIdParams) ([]DriftAnalysisRun, error) {
	rows, err := q.db.Query(ctx, findDriftAnalysisRunsByRepositoryId, arg.RepositoryID, arg.Queryoffset, arg.Maxresults)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []DriftAnalysisRun
	for rows.Next() {
		var i DriftAnalysisRun
		if err := rows.Scan(
			&i.Uuid,
			&i.RepositoryID,
			&i.TotalProjects,
			&i.TotalProjectsDrifted,
			&i.AnalysisDurationMillis,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
