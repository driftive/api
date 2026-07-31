package drift_stream

import (
	"context"
	"errors"
	"strings"

	"driftive.cloud/api/pkg/repository/queries"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// DriftProgressRequest is a partial report from a scan that is still in progress. Running holds
// the dirs currently being analyzed; ProjectResults holds whatever finished since the last tick
// and may be empty on a heartbeat.
type DriftProgressRequest struct {
	TotalProjects  int32                `json:"total_projects"`
	Running        []string             `json:"running"`
	ProjectResults []DriftProjectResult `json:"project_results"`
}

// HandleProgress records incremental progress for an in-flight run, creating the run on the first
// tick. The run is addressed by (repository_id, Idempotency-Key) rather than by a path param, so a
// token structurally cannot reach another repository's run.
func (d *DriftStateHandler) HandleProgress(c fiber.Ctx) error {
	repo, org, status, ok := d.resolveRepoAndOrg(c)
	if !ok {
		return c.SendStatus(status)
	}

	idemKey := strings.TrimSpace(c.Get("Idempotency-Key"))
	if idemKey == "" {
		log.Warnf("Rejecting drift progress for repository %d: missing Idempotency-Key", repo.ID)
		return c.SendStatus(fiber.StatusBadRequest)
	}

	var req DriftProgressRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	// pgx encodes a nil slice as NULL, which running_projects rejects.
	running := req.Running
	if running == nil {
		running = []string{}
	}

	run, err := d.findOrCreateRunningRun(c.Context(), repo.ID, idemKey, req.TotalProjects, running)
	if err != nil {
		log.Errorf("Error resolving running run for repository %d, key %s: %v", repo.ID, idemKey, err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// The finalize already landed; nothing to report onto it.
	if run.Status == runStatusCompleted {
		return c.JSON(buildAnalysisResponse(d.cfg.Frontend.FrontendURL, org, repo, run.Uuid))
	}

	upsertParams, err := toUpsertParams(run.Uuid, req.ProjectResults)
	if err != nil {
		log.Errorf("Rejecting drift progress: %v", err)
		return c.SendStatus(fiber.StatusBadRequest)
	}

	err = d.driftAnalysisRepository.WithTx(c.Context(), func(ctx context.Context) error {
		if len(upsertParams) > 0 {
			if err := d.driftAnalysisRepository.UpsertDriftAnalysisProjects(ctx, upsertParams); err != nil {
				log.Errorf("Error upserting drift analysis projects for run %s: %v", run.Uuid, err)
				return err
			}
		}
		// Recomputes the run counters, so it must not commit without the rows above.
		return d.driftAnalysisRepository.UpdateDriftAnalysisRunProgress(ctx, queries.UpdateDriftAnalysisRunProgressParams{
			Uuid:            run.Uuid,
			RunningProjects: running,
			TotalProjects:   req.TotalProjects,
		})
	})
	if err != nil {
		log.Errorf("Error recording drift progress for run %s: %v", run.Uuid, err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	log.Debugf("Recorded drift progress for run %s: %d result(s), %d running", run.Uuid, len(upsertParams), len(running))
	return c.JSON(buildAnalysisResponse(d.cfg.Frontend.FrontendURL, org, repo, run.Uuid))
}

// findOrCreateRunningRun returns the run for this idempotency key, creating it in the RUNNING
// state on the first tick. A concurrent creator trips the partial unique index on
// (repository_id, idempotency_key), in which case the winning row is re-fetched.
func (d *DriftStateHandler) findOrCreateRunningRun(
	ctx context.Context,
	repoID int64,
	idemKey string,
	totalProjects int32,
	running []string,
) (queries.DriftAnalysisRun, error) {
	run, err := d.driftAnalysisRepository.FindRunByRepoAndIdempotencyKey(ctx, repoID, idemKey)
	if err == nil {
		return run, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return queries.DriftAnalysisRun{}, err
	}

	created, err := d.driftAnalysisRepository.CreateRunningDriftAnalysisRun(ctx, queries.CreateRunningDriftAnalysisRunParams{
		Uuid:            uuid.New(),
		RepositoryID:    repoID,
		TotalProjects:   totalProjects,
		IdempotencyKey:  &idemKey,
		RunningProjects: running,
	})
	if err == nil {
		log.Infof("Started drift analysis run %s for repository %d, key %s", created.Uuid, repoID, idemKey)
		return created, nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
		return d.driftAnalysisRepository.FindRunByRepoAndIdempotencyKey(ctx, repoID, idemKey)
	}
	return queries.DriftAnalysisRun{}, err
}
