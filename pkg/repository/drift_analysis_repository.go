package repository

import (
	"context"
	"driftive.cloud/api/pkg/db"
	"driftive.cloud/api/pkg/repository/queries"
	"github.com/google/uuid"
)

type DriftAnalysisRepository interface {
	CreateDriftAnalysisRun(ctx context.Context, params queries.CreateDriftAnalysisRunParams) (queries.DriftAnalysisRun, error)
	CreateDriftAnalysisProject(ctx context.Context, params queries.CreateDriftAnalysisProjectParams) (queries.DriftAnalysisProject, error)
	FindDriftAnalysisRunsByRepositoryID(ctx context.Context, repoId int64, page int) ([]queries.DriftAnalysisRun, error)
	FindDriftAnalysisRunByUUID(ctx context.Context, uuid uuid.UUID) (queries.DriftAnalysisRun, error)
	FindDriftAnalysisProjectsByRunId(ctx context.Context, runId uuid.UUID) ([]queries.DriftAnalysisProject, error)

	WithTx(ctx context.Context, txFunc func(context.Context) error) error
}

type DriftAnalysisRepo struct {
	db *db.DB
}

func (r *DriftAnalysisRepo) CreateDriftAnalysisRun(ctx context.Context, params queries.CreateDriftAnalysisRunParams) (queries.DriftAnalysisRun, error) {
	return r.db.Queries(ctx).CreateDriftAnalysisRun(ctx, params)
}

func (r *DriftAnalysisRepo) CreateDriftAnalysisProject(ctx context.Context, params queries.CreateDriftAnalysisProjectParams) (queries.DriftAnalysisProject, error) {
	return r.db.Queries(ctx).CreateDriftAnalysisProject(ctx, params)
}

func (r *DriftAnalysisRepo) WithTx(ctx context.Context, txFunc func(context.Context) error) error {
	return r.db.WithTx(ctx, txFunc)
}

func (r *DriftAnalysisRepo) FindDriftAnalysisRunsByRepositoryID(ctx context.Context, repoId int64, page int) ([]queries.DriftAnalysisRun, error) {
	params := queries.FindDriftAnalysisRunsByRepositoryIdParams{
		RepositoryID: repoId,
		Queryoffset:  int32(page * 25),
		Maxresults:   25,
	}
	return r.db.Queries(ctx).FindDriftAnalysisRunsByRepositoryId(ctx, params)
}

func (r *DriftAnalysisRepo) FindDriftAnalysisRunByUUID(ctx context.Context, uuid uuid.UUID) (queries.DriftAnalysisRun, error) {
	return r.db.Queries(ctx).FindDriftAnalysisRunByUUID(ctx, uuid)
}

func (r *DriftAnalysisRepo) FindDriftAnalysisProjectsByRunId(ctx context.Context, runId uuid.UUID) ([]queries.DriftAnalysisProject, error) {
	return r.db.Queries(ctx).FindDriftAnalysisProjectsByRunId(ctx, runId)
}
