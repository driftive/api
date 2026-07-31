package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"driftive.cloud/api/pkg/config"
	"driftive.cloud/api/pkg/repository"
	"driftive.cloud/api/pkg/repository/queries"
	"driftive.cloud/api/pkg/usecase/drift_stream"
	"github.com/google/uuid"
)

// runRow is the subset of drift_analysis_run the progress tests assert on. Kept comparable so a
// "nothing changed" assertion can be a single ==; running_projects is fetched separately.
type runRow struct {
	uuid           string
	status         string
	totalProjects  int32
	drifted        int32
	errored        int32
	skipped        int32
	durationMillis int64
}

func newDriftRepo(t *testing.T) repository.DriftAnalysisRepository {
	t.Helper()
	if testDB == nil {
		t.Skip("integration tests skipped (no testDB)")
	}
	repos := repository.NewRepository(testDB, &config.Config{})
	return repos.DriftAnalysisRepository()
}

func fetchRun(t *testing.T, repoID int64) (runRow, []string) {
	t.Helper()
	var r runRow
	var runningProjects []string
	err := withPool(t).QueryRow(context.Background(),
		`SELECT uuid, status, running_projects, total_projects, total_projects_drifted,
		        total_projects_errored, total_projects_skipped, analysis_duration_millis
		 FROM drift_analysis_run WHERE repository_id = $1`, repoID).
		Scan(&r.uuid, &r.status, &runningProjects, &r.totalProjects, &r.drifted,
			&r.errored, &r.skipped, &r.durationMillis)
	if err != nil {
		t.Fatalf("fetch run: %v", err)
	}
	return r, runningProjects
}

func countRuns(t *testing.T) int {
	t.Helper()
	var n int
	if err := withPool(t).QueryRow(context.Background(),
		`SELECT count(*) FROM drift_analysis_run`).Scan(&n); err != nil {
		t.Fatalf("count runs: %v", err)
	}
	return n
}

// projectByDir returns the id and plan_output of one project row, so the tests can prove the
// upsert updates in place instead of inserting a second row.
func projectByDir(t *testing.T, runUUID, dir string) (int64, string) {
	t.Helper()
	var id int64
	var planOutput *string
	err := withPool(t).QueryRow(context.Background(),
		`SELECT id, plan_output FROM drift_analysis_project
		 WHERE drift_analysis_run_id = $1::uuid AND dir = $2`, runUUID, dir).Scan(&id, &planOutput)
	if err != nil {
		t.Fatalf("fetch project %s: %v", dir, err)
	}
	if planOutput == nil {
		return id, ""
	}
	return id, *planOutput
}

func countProjects(t *testing.T, runUUID string) int {
	t.Helper()
	var n int
	if err := withPool(t).QueryRow(context.Background(),
		`SELECT count(*) FROM drift_analysis_project WHERE drift_analysis_run_id = $1::uuid`,
		runUUID).Scan(&n); err != nil {
		t.Fatalf("count projects: %v", err)
	}
	return n
}

func runIDFromResponse(t *testing.T, body []byte) string {
	t.Helper()
	var r struct {
		RunID        string `json:"run_id"`
		DashboardURL string `json:"dashboard_url"`
	}
	if err := json.Unmarshal(body, &r); err != nil {
		t.Fatalf("decode response %s: %v", body, err)
	}
	if r.RunID == "" {
		t.Fatalf("response missing run_id: %s", body)
	}
	if r.DashboardURL == "" {
		t.Fatalf("response missing dashboard_url: %s", body)
	}
	return r.RunID
}

func driftedProject(dir, planOutput string) drift_stream.DriftProjectResult {
	return drift_stream.DriftProjectResult{
		Project:    drift_stream.TypedProject{Dir: dir, Type: drift_stream.Terraform},
		Drifted:    true,
		Succeeded:  true,
		InitOutput: "init-" + dir,
		PlanOutput: planOutput,
	}
}

// TestProgress_CreatesRunAndUpsertsWithoutDuplicates is the core guarantee of live reporting: the
// first tick creates the RUNNING run, and re-sending a project updates its row rather than adding
// a second one.
func TestProgress_CreatesRunAndUpsertsWithoutDuplicates(t *testing.T) {
	truncateAll(t)
	repoID := seedOrgAndRepo(t)
	app := newIngestApp(t)

	const idemKey = "progress-key-1"

	status, body := postProgress(t, app, seedAnalysisToken, idemKey, drift_stream.DriftProgressRequest{
		TotalProjects:  3,
		Running:        []string{"projects/a", "projects/b"},
		ProjectResults: nil,
	})
	if status != http.StatusOK {
		t.Fatalf("first progress: expected 200, got %d: %s", status, body)
	}
	firstRunID := runIDFromResponse(t, body)

	run, running := fetchRun(t, repoID)
	if run.status != "RUNNING" {
		t.Errorf("expected status RUNNING, got %q", run.status)
	}
	if run.totalProjects != 3 {
		t.Errorf("expected total_projects 3, got %d", run.totalProjects)
	}
	if len(running) != 2 {
		t.Errorf("expected 2 running projects, got %v", running)
	}

	// Second tick: project a finishes with a short plan, b is still running.
	status, body = postProgress(t, app, seedAnalysisToken, idemKey, drift_stream.DriftProgressRequest{
		TotalProjects: 3,
		Running:       []string{"projects/b"},
		ProjectResults: []drift_stream.DriftProjectResult{
			driftedProject("projects/a", "short plan"),
		},
	})
	if status != http.StatusOK {
		t.Fatalf("second progress: expected 200, got %d: %s", status, body)
	}
	if got := runIDFromResponse(t, body); got != firstRunID {
		t.Fatalf("expected the same run id across ticks, got %s then %s", firstRunID, got)
	}
	firstID, _ := projectByDir(t, firstRunID, "projects/a")

	// Third tick: a is re-sent with a fuller plan, and c arrives errored.
	status, body = postProgress(t, app, seedAnalysisToken, idemKey, drift_stream.DriftProgressRequest{
		TotalProjects: 3,
		Running:       []string{},
		ProjectResults: []drift_stream.DriftProjectResult{
			driftedProject("projects/a", "Plan: 2 to add, 1 to change, 3 to destroy"),
			{
				Project:   drift_stream.TypedProject{Dir: "projects/c", Type: drift_stream.Tofu},
				Drifted:   false,
				Succeeded: false,
			},
		},
	})
	if status != http.StatusOK {
		t.Fatalf("third progress: expected 200, got %d: %s", status, body)
	}

	if n := countRuns(t); n != 1 {
		t.Errorf("expected exactly 1 run across three ticks, got %d", n)
	}
	if n := countProjects(t, firstRunID); n != 2 {
		t.Errorf("expected 2 project rows, got %d", n)
	}

	secondID, planOutput := projectByDir(t, firstRunID, "projects/a")
	if secondID != firstID {
		t.Errorf("upsert replaced the row (id %d -> %d) instead of updating in place", firstID, secondID)
	}
	if planOutput != "Plan: 2 to add, 1 to change, 3 to destroy" {
		t.Errorf("expected plan_output to be updated, got %q", planOutput)
	}

	// The re-sent plan summary must be parsed on the update, not only on the insert.
	var added, changed, destroyed *int32
	if err := withPool(t).QueryRow(context.Background(),
		`SELECT resources_added, resources_changed, resources_destroyed
		 FROM drift_analysis_project WHERE id = $1`, secondID).
		Scan(&added, &changed, &destroyed); err != nil {
		t.Fatalf("query resource counts: %v", err)
	}
	if added == nil || changed == nil || destroyed == nil {
		t.Fatalf("expected parsed resource counts, got %v/%v/%v", added, changed, destroyed)
	}
	if *added != 2 || *changed != 1 || *destroyed != 3 {
		t.Errorf("resource counts = %d/%d/%d, want 2/1/3", *added, *changed, *destroyed)
	}

	run, running = fetchRun(t, repoID)
	if run.status != "RUNNING" {
		t.Errorf("run should still be RUNNING, got %q", run.status)
	}
	if len(running) != 0 {
		t.Errorf("expected running_projects to be replaced with empty, got %v", running)
	}
	// Recomputed from the project rows: a drifted, c errored.
	if run.drifted != 1 || run.errored != 1 || run.skipped != 0 {
		t.Errorf("recomputed counters = %d drifted / %d errored / %d skipped, want 1/1/0",
			run.drifted, run.errored, run.skipped)
	}
}

// TestFinalizeCompletesRunningRun covers the handoff: the terminal ingest adopts the RUNNING run
// created by progress rather than starting a second one.
func TestFinalizeCompletesRunningRun(t *testing.T) {
	truncateAll(t)
	repoID := seedOrgAndRepo(t)
	app := newIngestApp(t)

	const idemKey = "handoff-key"

	status, body := postProgress(t, app, seedAnalysisToken, idemKey, drift_stream.DriftProgressRequest{
		TotalProjects: 3,
		Running:       []string{"/projects/a"},
		ProjectResults: []drift_stream.DriftProjectResult{
			driftedProject("/projects/a", "partial plan"),
		},
	})
	if status != http.StatusOK {
		t.Fatalf("progress: expected 200, got %d: %s", status, body)
	}
	runID := runIDFromResponse(t, body)
	projectAID, _ := projectByDir(t, runID, "/projects/a")

	status, body = postIngest(t, app, seedAnalysisToken, idemKey, sampleState())
	if status != http.StatusOK {
		t.Fatalf("finalize: expected 200, got %d: %s", status, body)
	}
	if got := runIDFromResponse(t, body); got != runID {
		t.Fatalf("finalize started a new run %s instead of adopting %s", got, runID)
	}

	if n := countRuns(t); n != 1 {
		t.Errorf("expected 1 run after finalize, got %d", n)
	}
	if n := countProjects(t, runID); n != 3 {
		t.Errorf("expected 3 project rows from the payload, got %d", n)
	}
	if id, _ := projectByDir(t, runID, "/projects/a"); id != projectAID {
		t.Errorf("finalize replaced project a (id %d -> %d) instead of updating it", projectAID, id)
	}

	run, running := fetchRun(t, repoID)
	if run.status != "COMPLETED" {
		t.Errorf("expected status COMPLETED, got %q", run.status)
	}
	if len(running) != 0 {
		t.Errorf("expected running_projects cleared, got %v", running)
	}
	if run.totalProjects != 3 || run.drifted != 1 || run.skipped != 1 || run.errored != 0 {
		t.Errorf("totals = %d projects / %d drifted / %d errored / %d skipped, want 3/1/0/1",
			run.totalProjects, run.drifted, run.errored, run.skipped)
	}
	if run.durationMillis != 250 {
		t.Errorf("expected duration 250ms from the payload, got %d", run.durationMillis)
	}
}

// TestProgressAfterCompletionIsNoop pins the late-tick case: a progress post that arrives after
// the finalize must not reopen or mutate the completed run.
func TestProgressAfterCompletionIsNoop(t *testing.T) {
	truncateAll(t)
	repoID := seedOrgAndRepo(t)
	app := newIngestApp(t)

	const idemKey = "late-tick-key"

	status, body := postIngest(t, app, seedAnalysisToken, idemKey, sampleState())
	if status != http.StatusOK {
		t.Fatalf("finalize: expected 200, got %d: %s", status, body)
	}
	runID := runIDFromResponse(t, body)
	before, _ := fetchRun(t, repoID)

	status, body = postProgress(t, app, seedAnalysisToken, idemKey, drift_stream.DriftProgressRequest{
		TotalProjects: 99,
		Running:       []string{"projects/zombie"},
		ProjectResults: []drift_stream.DriftProjectResult{
			driftedProject("projects/zombie", "should not land"),
		},
	})
	if status != http.StatusOK {
		t.Fatalf("late progress: expected 200, got %d: %s", status, body)
	}
	if got := runIDFromResponse(t, body); got != runID {
		t.Errorf("late progress returned run %s, want %s", got, runID)
	}

	after, running := fetchRun(t, repoID)
	if after != before {
		t.Errorf("late progress mutated the completed run:\n before=%+v\n after=%+v", before, after)
	}
	if len(running) != 0 {
		t.Errorf("late progress set running_projects on a completed run: %v", running)
	}
	if n := countProjects(t, runID); n != 3 {
		t.Errorf("late progress added project rows: expected 3, got %d", n)
	}
}

// TestProgressUpdateGuardsOnRunningStatus exercises the AND status = 'RUNNING' guard in
// UpdateDriftAnalysisRunProgress directly. HandleProgress returns early on a completed run, so
// this backstop — the one that matters if a tick ever races the finalize — is otherwise untested.
func TestProgressUpdateGuardsOnRunningStatus(t *testing.T) {
	truncateAll(t)
	repoID := seedOrgAndRepo(t)
	app := newIngestApp(t)
	repo := newDriftRepo(t)
	ctx := context.Background()

	if status, body := postIngest(t, app, seedAnalysisToken, "guard-key", sampleState()); status != http.StatusOK {
		t.Fatalf("finalize: expected 200, got %d: %s", status, body)
	}
	before, _ := fetchRun(t, repoID)

	runUUID, err := uuid.Parse(before.uuid)
	if err != nil {
		t.Fatalf("parse run uuid: %v", err)
	}
	err = repo.UpdateDriftAnalysisRunProgress(ctx, queries.UpdateDriftAnalysisRunProgressParams{
		Uuid:            runUUID,
		RunningProjects: []string{"projects/racing"},
		TotalProjects:   999,
	})
	if err != nil {
		t.Fatalf("UpdateDriftAnalysisRunProgress: %v", err)
	}

	after, running := fetchRun(t, repoID)
	if after != before {
		t.Errorf("progress update mutated a COMPLETED run:\n before=%+v\n after=%+v", before, after)
	}
	if len(running) != 0 {
		t.Errorf("progress update set running_projects on a COMPLETED run: %v", running)
	}
}

func TestProgress_InvalidToken(t *testing.T) {
	truncateAll(t)
	seedOrgAndRepo(t)
	app := newIngestApp(t)

	status, _ := postProgress(t, app, "no-such-token", "any-key", drift_stream.DriftProgressRequest{TotalProjects: 1})
	if status != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", status)
	}
	if n := countRuns(t); n != 0 {
		t.Errorf("expected no run to be created, got %d", n)
	}
}

func TestProgress_MissingIdempotencyKey(t *testing.T) {
	truncateAll(t)
	seedOrgAndRepo(t)
	app := newIngestApp(t)

	status, _ := postProgress(t, app, seedAnalysisToken, "", drift_stream.DriftProgressRequest{TotalProjects: 1})
	if status != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", status)
	}
	if n := countRuns(t); n != 0 {
		t.Errorf("expected no run to be created without a key, got %d", n)
	}
}

// TestProgress_RunningRunIsExcludedFromAggregates verifies the status filter added to the
// aggregate reads: an in-flight run must not become the repository's "latest run", but must still
// appear in the runs list so users can click into it.
func TestProgress_RunningRunIsExcludedFromAggregates(t *testing.T) {
	truncateAll(t)
	repoID := seedOrgAndRepo(t)
	app := newIngestApp(t)
	ctx := context.Background()

	if status, body := postIngest(t, app, seedAnalysisToken, "completed-run", sampleState()); status != http.StatusOK {
		t.Fatalf("finalize: expected 200, got %d: %s", status, body)
	}
	var completedUUID string
	if err := withPool(t).QueryRow(ctx,
		`SELECT uuid FROM drift_analysis_run WHERE repository_id = $1`, repoID).Scan(&completedUUID); err != nil {
		t.Fatalf("fetch completed run: %v", err)
	}

	// A newer RUNNING run would otherwise win the ORDER BY created_at DESC.
	status, body := postProgress(t, app, seedAnalysisToken, "live-run", drift_stream.DriftProgressRequest{
		TotalProjects: 5,
		Running:       []string{"projects/live"},
	})
	if status != http.StatusOK {
		t.Fatalf("progress: expected 200, got %d: %s", status, body)
	}

	repo := newDriftRepo(t)
	latest, err := repo.GetLatestRunForRepository(ctx, repoID)
	if err != nil {
		t.Fatalf("GetLatestRunForRepository: %v", err)
	}
	if latest.Uuid.String() != completedUUID {
		t.Errorf("latest run = %s, want the completed run %s", latest.Uuid, completedUUID)
	}

	rate, err := repo.GetDriftRateOverTime(ctx, repoID, 30)
	if err != nil {
		t.Fatalf("GetDriftRateOverTime: %v", err)
	}
	var totalRuns int64
	for _, dp := range rate {
		totalRuns += dp.TotalRuns
	}
	if totalRuns != 1 {
		t.Errorf("drift rate counted %d runs, want only the 1 completed run", totalRuns)
	}

	runs, err := repo.FindDriftAnalysisRunsByRepositoryID(ctx, repoID, 0)
	if err != nil {
		t.Fatalf("FindDriftAnalysisRunsByRepositoryID: %v", err)
	}
	if len(runs) != 2 {
		t.Errorf("expected the runs list to include the live run (2 total), got %d", len(runs))
	}
}
