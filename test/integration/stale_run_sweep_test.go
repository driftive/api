package integration

import (
	"context"
	"fmt"
	"sync"
	"testing"
)

const sweepStaleMinutes = 15

// seedRun inserts a run with an explicit status and updated_at age, so the sweeper's threshold can
// be exercised without waiting.
func seedRun(t *testing.T, repoID int64, status string, ageMinutes int, idemKey string) string {
	t.Helper()
	var runUUID string
	err := withPool(t).QueryRow(context.Background(),
		`INSERT INTO drift_analysis_run
		     (uuid, repository_id, total_projects, total_projects_drifted, total_projects_errored,
		      total_projects_skipped, analysis_duration_millis, idempotency_key, status, updated_at)
		 VALUES (gen_random_uuid(), $1, 0, 0, 0, 0, 0, $2, $3,
		         NOW() - ($4::INTEGER || ' minutes')::INTERVAL)
		 RETURNING uuid`, repoID, idemKey, status, ageMinutes).Scan(&runUUID)
	if err != nil {
		t.Fatalf("seed %s run: %v", status, err)
	}
	return runUUID
}

func seedProject(t *testing.T, runUUID, dir string) {
	t.Helper()
	_, err := withPool(t).Exec(context.Background(),
		`INSERT INTO drift_analysis_project (drift_analysis_run_id, dir, type, drifted, succeeded)
		 VALUES ($1::uuid, $2, 'TERRAFORM', true, true)`, runUUID, dir)
	if err != nil {
		t.Fatalf("seed project %s: %v", dir, err)
	}
}

func runExists(t *testing.T, runUUID string) bool {
	t.Helper()
	var exists bool
	if err := withPool(t).QueryRow(context.Background(),
		`SELECT EXISTS (SELECT 1 FROM drift_analysis_run WHERE uuid = $1::uuid)`, runUUID).
		Scan(&exists); err != nil {
		t.Fatalf("check run %s: %v", runUUID, err)
	}
	return exists
}

// TestSweepDeletesOnlyStaleRunningRuns pins the threshold and the status guard: a fresh RUNNING
// run and an old COMPLETED run must both survive.
func TestSweepDeletesOnlyStaleRunningRuns(t *testing.T) {
	truncateAll(t)
	repoID := seedOrgAndRepo(t)
	repo := newDriftRepo(t)
	ctx := context.Background()

	staleRunning := seedRun(t, repoID, "RUNNING", 20, "stale-running")
	seedProject(t, staleRunning, "projects/a")
	seedProject(t, staleRunning, "projects/b")

	freshRunning := seedRun(t, repoID, "RUNNING", 1, "fresh-running")
	oldCompleted := seedRun(t, repoID, "COMPLETED", 20, "old-completed")

	deleted, err := repo.DeleteStaleRunningRuns(ctx, sweepStaleMinutes, 100)
	if err != nil {
		t.Fatalf("DeleteStaleRunningRuns: %v", err)
	}
	if deleted != 1 {
		t.Errorf("expected 1 run swept, got %d", deleted)
	}

	if runExists(t, staleRunning) {
		t.Error("the stale RUNNING run should have been deleted")
	}
	if !runExists(t, freshRunning) {
		t.Error("a RUNNING run inside the threshold must survive")
	}
	if !runExists(t, oldCompleted) {
		t.Error("an old COMPLETED run must survive")
	}

	// Project rows go with the run via ON DELETE CASCADE.
	if n := countProjects(t, staleRunning); n != 0 {
		t.Errorf("expected the swept run's project rows to be cascaded away, %d remain", n)
	}
}

// TestSweepIsSafeConcurrently is the FOR UPDATE SKIP LOCKED assertion: two sweepers racing (as two
// API instances would) must not error and must delete the run exactly once between them.
func TestSweepIsSafeConcurrently(t *testing.T) {
	truncateAll(t)
	repoID := seedOrgAndRepo(t)
	repo := newDriftRepo(t)

	staleRunning := seedRun(t, repoID, "RUNNING", 30, "raced-run")
	seedProject(t, staleRunning, "projects/a")

	const sweepers = 4
	var wg sync.WaitGroup
	deletedCounts := make([]int64, sweepers)
	errs := make([]error, sweepers)
	for i := 0; i < sweepers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			deletedCounts[idx], errs[idx] = repo.DeleteStaleRunningRuns(context.Background(), sweepStaleMinutes, 100)
		}(i)
	}
	wg.Wait()

	var total int64
	for i, err := range errs {
		if err != nil {
			t.Errorf("sweeper %d errored: %v", i, err)
		}
		total += deletedCounts[i]
	}
	if total != 1 {
		t.Errorf("concurrent sweepers deleted %d rows in total, want exactly 1", total)
	}
	if runExists(t, staleRunning) {
		t.Error("the stale run should be gone after the concurrent sweep")
	}
}

// TestSweepRespectsBatchCap verifies the LIMIT bounds one statement and that the remainder is
// picked up by the next call rather than silently dropped.
func TestSweepRespectsBatchCap(t *testing.T) {
	truncateAll(t)
	repoID := seedOrgAndRepo(t)
	repo := newDriftRepo(t)
	ctx := context.Background()

	const batchCap = 5
	const extra = 3
	for i := 0; i < batchCap+extra; i++ {
		seedRun(t, repoID, "RUNNING", 20, fmt.Sprintf("stale-%d", i))
	}

	deleted, err := repo.DeleteStaleRunningRuns(ctx, sweepStaleMinutes, batchCap)
	if err != nil {
		t.Fatalf("first sweep: %v", err)
	}
	if deleted != batchCap {
		t.Errorf("first sweep deleted %d, want the cap of %d", deleted, batchCap)
	}
	if n := countRuns(t); n != extra {
		t.Errorf("expected %d runs left after the capped sweep, got %d", extra, n)
	}

	deleted, err = repo.DeleteStaleRunningRuns(ctx, sweepStaleMinutes, batchCap)
	if err != nil {
		t.Fatalf("second sweep: %v", err)
	}
	if deleted != extra {
		t.Errorf("second sweep deleted %d, want the remaining %d", deleted, extra)
	}
	if n := countRuns(t); n != 0 {
		t.Errorf("expected the sweep to drain, %d runs left", n)
	}
}
