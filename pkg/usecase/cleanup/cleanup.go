package cleanup

import (
	"context"
	"time"

	"driftive.cloud/api/pkg/repository"
	"github.com/gofiber/fiber/v3/log"
)

const (
	staleRunSweepInterval = 5 * time.Minute
	staleRunSweepBatch    = 100
)

type CleanupService struct {
	driftAnalysisRepo repository.DriftAnalysisRepository
	maxRunsPerRepo    int32
	staleRunMinutes   int32
}

func NewCleanupService(driftAnalysisRepo repository.DriftAnalysisRepository, maxRunsPerRepo int32, staleRunMinutes int32) *CleanupService {
	return &CleanupService{
		driftAnalysisRepo: driftAnalysisRepo,
		maxRunsPerRepo:    maxRunsPerRepo,
		staleRunMinutes:   staleRunMinutes,
	}
}

// CleanupRepositoryRuns deletes the oldest runs for a repository, keeping only the most recent N runs.
func (s *CleanupService) CleanupRepositoryRuns(ctx context.Context, repoId int64) error {
	return s.driftAnalysisRepo.DeleteOldestRunsExceedingLimit(ctx, repoId, s.maxRunsPerRepo)
}

// StartStaleRunSweeper deletes runs a crashed CLI left in the RUNNING state. Safe to run on every
// API instance at once: DeleteStaleRunningRuns claims rows with FOR UPDATE SKIP LOCKED, so
// concurrent sweepers neither block each other nor delete the same run twice.
func (s *CleanupService) StartStaleRunSweeper(ctx context.Context) {
	for {
		// Sleeping first keeps a fleet-wide deploy from sweeping all at once on boot.
		select {
		case <-ctx.Done():
			log.Info("stale run sweeper shutting down...")
			return
		case <-time.After(staleRunSweepInterval):
		}

		deleted, err := s.driftAnalysisRepo.DeleteStaleRunningRuns(ctx, s.staleRunMinutes, staleRunSweepBatch)
		switch {
		case err != nil:
			log.Errorf("error sweeping stale running runs: %v", err)
		case deleted >= staleRunSweepBatch:
			log.Warnf("swept %d stale running run(s); batch cap reached, more may remain", deleted)
		case deleted > 0:
			log.Infof("swept %d stale running run(s)", deleted)
		}
	}
}
