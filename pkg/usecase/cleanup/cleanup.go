package cleanup

import (
	"context"

	"driftive.cloud/api/pkg/repository"
)

type CleanupService struct {
	driftAnalysisRepo repository.DriftAnalysisRepository
	maxRunsPerRepo    int32
}

func NewCleanupService(driftAnalysisRepo repository.DriftAnalysisRepository, maxRunsPerRepo int32) *CleanupService {
	return &CleanupService{
		driftAnalysisRepo: driftAnalysisRepo,
		maxRunsPerRepo:    maxRunsPerRepo,
	}
}

// CleanupRepositoryRuns deletes the oldest runs for a repository, keeping only the most recent N runs.
func (s *CleanupService) CleanupRepositoryRuns(ctx context.Context, repoId int64) error {
	return s.driftAnalysisRepo.DeleteOldestRunsExceedingLimit(ctx, repoId, s.maxRunsPerRepo)
}
