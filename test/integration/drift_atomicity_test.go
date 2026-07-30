package integration

import (
	"context"
	"strings"
	"testing"
	"time"

	"driftive.cloud/api/pkg/usecase/drift_stream"
)

// TestDriftIngest_RunAndProjectsAreAtomic is the regression test for the db.WithTx context-key
// bug: WithTx stored the transaction under TxType("tx") while Queries read it back under the
// untyped string "tx", so every statement inside the closure ran on the pool and autocommitted.
// The run INSERT therefore survived a failing project batch.
//
// Here the third project's dir exceeds drift_analysis_project.dir VARCHAR(1500), so the batch
// insert fails. Nothing from the request may survive.
func TestDriftIngest_RunAndProjectsAreAtomic(t *testing.T) {
	truncateAll(t)
	seedOrgAndRepo(t)
	app := newIngestApp(t)
	ctx := context.Background()
	pool := withPool(t)

	totalErrored := int32(0)
	state := drift_stream.DriftDetectionResult{
		ProjectResults: []drift_stream.DriftProjectResult{
			{
				Project:   drift_stream.TypedProject{Dir: "projects/a", Type: drift_stream.Terraform},
				Drifted:   true,
				Succeeded: true,
			},
			{
				Project:   drift_stream.TypedProject{Dir: strings.Repeat("x", 1600), Type: drift_stream.Terraform},
				Drifted:   false,
				Succeeded: true,
			},
		},
		TotalDrifted:  1,
		TotalErrored:  &totalErrored,
		TotalProjects: 2,
		TotalChecked:  2,
		Duration:      100 * time.Millisecond,
	}

	status, _ := postIngest(t, app, seedAnalysisToken, "atomicity-key", state)
	if status != 500 {
		t.Fatalf("expected 500 from the failing batch insert, got %d", status)
	}

	var runCount int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM drift_analysis_run`).Scan(&runCount); err != nil {
		t.Fatalf("count runs: %v", err)
	}
	if runCount != 0 {
		t.Errorf("expected the run insert to roll back with the failing batch, found %d orphan run(s)", runCount)
	}

	var projectCount int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM drift_analysis_project`).Scan(&projectCount); err != nil {
		t.Fatalf("count projects: %v", err)
	}
	if projectCount != 0 {
		t.Errorf("expected no project rows, found %d", projectCount)
	}
}

// TestDriftIngest_RetryAfterFailedBatchPersistsResults covers the narrowed idempotency-key
// early return. A first attempt that leaves a run with no project rows must be adopted by the
// retry rather than replayed, otherwise the results are swallowed and never stored.
func TestDriftIngest_RetryAfterFailedBatchPersistsResults(t *testing.T) {
	truncateAll(t)
	repoID := seedOrgAndRepo(t)
	app := newIngestApp(t)
	ctx := context.Background()
	pool := withPool(t)

	const idemKey = "adopt-me"

	// Simulate the aftermath of a partial write: a run row under the idempotency key with no
	// project rows attached.
	var runUUID string
	err := pool.QueryRow(ctx,
		`INSERT INTO drift_analysis_run
		     (uuid, repository_id, total_projects, total_projects_drifted, total_projects_errored,
		      total_projects_skipped, analysis_duration_millis, idempotency_key)
		 VALUES (gen_random_uuid(), $1, 0, 0, 0, 0, 0, $2)
		 RETURNING uuid`, repoID, idemKey).Scan(&runUUID)
	if err != nil {
		t.Fatalf("seed empty run: %v", err)
	}

	status, _ := postIngest(t, app, seedAnalysisToken, idemKey, sampleState())
	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}

	// The empty run must have been adopted, not duplicated.
	var runCount int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM drift_analysis_run`).Scan(&runCount); err != nil {
		t.Fatalf("count runs: %v", err)
	}
	if runCount != 1 {
		t.Errorf("expected the existing run to be adopted, got %d runs", runCount)
	}

	var projectCount int
	err = pool.QueryRow(ctx,
		`SELECT count(*) FROM drift_analysis_project WHERE drift_analysis_run_id = $1::uuid`,
		runUUID).Scan(&projectCount)
	if err != nil {
		t.Fatalf("count projects: %v", err)
	}
	if projectCount != 3 {
		t.Errorf("expected the 3 payload projects to be written to the adopted run, got %d", projectCount)
	}
}

// TestDriftIngest_ReplayWithResultsIsNoop pins the other half of that condition: once a run
// holds results, a re-send under the same key must stay a no-op.
func TestDriftIngest_ReplayWithResultsIsNoop(t *testing.T) {
	truncateAll(t)
	seedOrgAndRepo(t)
	app := newIngestApp(t)
	ctx := context.Background()
	pool := withPool(t)

	const idemKey = "replay-key"

	if status, _ := postIngest(t, app, seedAnalysisToken, idemKey, sampleState()); status != 200 {
		t.Fatalf("first post: expected 200, got %d", status)
	}
	if status, _ := postIngest(t, app, seedAnalysisToken, idemKey, sampleState()); status != 200 {
		t.Fatalf("replay: expected 200, got %d", status)
	}

	var runCount, projectCount int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM drift_analysis_run`).Scan(&runCount); err != nil {
		t.Fatalf("count runs: %v", err)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM drift_analysis_project`).Scan(&projectCount); err != nil {
		t.Fatalf("count projects: %v", err)
	}
	if runCount != 1 {
		t.Errorf("expected 1 run after replay, got %d", runCount)
	}
	if projectCount != 3 {
		t.Errorf("expected 3 projects after replay, got %d", projectCount)
	}
}
