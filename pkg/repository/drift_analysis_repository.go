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
	GetRepositoryRunStats(ctx context.Context, repoId int64) (queries.GetRepositoryRunStatsRow, error)
	GetLatestRunForRepository(ctx context.Context, repoId int64) (queries.DriftAnalysisRun, error)

	// Trend analytics methods
	GetDriftRateOverTime(ctx context.Context, repoId int64, daysBack int32) ([]queries.GetDriftRateOverTimeRow, error)
	GetMostFrequentlyDriftedProjects(ctx context.Context, repoId int64, daysBack int32, maxResults int32) ([]queries.GetMostFrequentlyDriftedProjectsRow, error)
	GetDriftFreeStreak(ctx context.Context, repoId int64) (queries.GetDriftFreeStreakRow, error)
	GetMeanTimeToResolution(ctx context.Context, repoId int64, daysBack int32) ([]queries.GetMeanTimeToResolutionRow, error)

	// Cleanup methods
	DeleteOldestRunsExceedingLimit(ctx context.Context, repoId int64, maxRunsToKeep int32) error

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

func (r *DriftAnalysisRepo) GetRepositoryRunStats(ctx context.Context, repoId int64) (queries.GetRepositoryRunStatsRow, error) {
	return r.db.Queries(ctx).GetRepositoryRunStats(ctx, repoId)
}

func (r *DriftAnalysisRepo) GetLatestRunForRepository(ctx context.Context, repoId int64) (queries.DriftAnalysisRun, error) {
	return r.db.Queries(ctx).GetLatestRunForRepository(ctx, repoId)
}

func (r *DriftAnalysisRepo) WithTx(ctx context.Context, txFunc func(context.Context) error) error {
	return r.db.WithTx(ctx, txFunc)
}

func (r *DriftAnalysisRepo) GetDriftRateOverTime(ctx context.Context, repoId int64, daysBack int32) ([]queries.GetDriftRateOverTimeRow, error) {
	return r.db.Queries(ctx).GetDriftRateOverTime(ctx, queries.GetDriftRateOverTimeParams{
		RepositoryID: repoId,
		DaysBack:     daysBack,
	})
}

func (r *DriftAnalysisRepo) GetMostFrequentlyDriftedProjects(ctx context.Context, repoId int64, daysBack int32, maxResults int32) ([]queries.GetMostFrequentlyDriftedProjectsRow, error) {
	return r.db.Queries(ctx).GetMostFrequentlyDriftedProjects(ctx, queries.GetMostFrequentlyDriftedProjectsParams{
		RepositoryID: repoId,
		DaysBack:     daysBack,
		MaxResults:   maxResults,
	})
}

func (r *DriftAnalysisRepo) GetDriftFreeStreak(ctx context.Context, repoId int64) (queries.GetDriftFreeStreakRow, error) {
	return r.db.Queries(ctx).GetDriftFreeStreak(ctx, repoId)
}

func (r *DriftAnalysisRepo) GetMeanTimeToResolution(ctx context.Context, repoId int64, daysBack int32) ([]queries.GetMeanTimeToResolutionRow, error) {
	return r.db.Queries(ctx).GetMeanTimeToResolution(ctx, queries.GetMeanTimeToResolutionParams{
		RepositoryID: repoId,
		DaysBack:     daysBack,
	})
}

func (r *DriftAnalysisRepo) DeleteOldestRunsExceedingLimit(ctx context.Context, repoId int64, maxRunsToKeep int32) error {
	return r.db.Queries(ctx).DeleteOldestRunsExceedingLimit(ctx, queries.DeleteOldestRunsExceedingLimitParams{
		RepositoryID:  repoId,
		MaxRunsToKeep: maxRunsToKeep,
	})
}
