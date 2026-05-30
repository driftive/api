package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"driftive.cloud/api/pkg/config"
	"driftive.cloud/api/pkg/repository"
	"driftive.cloud/api/pkg/usecase/cleanup"
	"driftive.cloud/api/pkg/usecase/drift_stream"
	"github.com/gofiber/fiber/v3"
)

const (
	seedOrgName        = "acme"
	seedProvider       = "GITHUB"
	seedRepoName       = "infra"
	seedAnalysisToken  = "test-analysis-token-abc"
	seedProviderOrgId  = "555"
	seedProviderRepoId = "777"
)

// seedOrgAndRepo inserts one org + one repo with a known analysis token and
// returns the inserted repo ID.
func seedOrgAndRepo(t *testing.T) (repoID int64) {
	t.Helper()
	ctx := context.Background()
	pool := withPool(t)

	var orgID int64
	err := pool.QueryRow(ctx,
		`INSERT INTO git_organization (provider, provider_id, name)
		 VALUES ($1, $2, $3) RETURNING id`,
		seedProvider, seedProviderOrgId, seedOrgName).Scan(&orgID)
	if err != nil {
		t.Fatalf("seed org: %v", err)
	}

	err = pool.QueryRow(ctx,
		`INSERT INTO git_repository (organization_id, provider_id, name, is_private, analysis_token)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		orgID, seedProviderRepoId, seedRepoName, false, seedAnalysisToken).Scan(&repoID)
	if err != nil {
		t.Fatalf("seed repo: %v", err)
	}
	return repoID
}

// newIngestApp builds a minimal Fiber app exposing just the drift ingest endpoint
// against the shared testDB. Mirrors the public-route registration in main.go.
func newIngestApp(t *testing.T) *fiber.App {
	t.Helper()
	repos := repository.NewRepository(testDB, &config.Config{})
	cleanupSvc := cleanup.NewCleanupService(repos.DriftAnalysisRepository(), 400)
	cfg := &config.Config{
		Frontend: config.FrontendConfig{FrontendURL: "http://test.local"},
	}
	handler := drift_stream.NewDriftStateHandler(
		cfg,
		repos.GitOrgRepository(),
		repos.GitRepoRepository(),
		repos.DriftAnalysisRepository(),
		cleanupSvc,
	)
	app := fiber.New()
	app.Post("/api/v1/drift_analysis", func(c fiber.Ctx) error { return handler.HandleUpdate(c) })
	return app
}

func sampleState() drift_stream.DriftDetectionResult {
	totalErrored := int32(0)
	return drift_stream.DriftDetectionResult{
		ProjectResults: []drift_stream.DriftProjectResult{
			{
				Project:    drift_stream.TypedProject{Dir: "/projects/a", Type: drift_stream.Terraform},
				Drifted:    true,
				Succeeded:  true,
				InitOutput: "init-a",
				PlanOutput: "plan-a",
			},
			{
				Project:    drift_stream.TypedProject{Dir: "/projects/b", Type: drift_stream.Tofu},
				Drifted:    false,
				Succeeded:  true,
				InitOutput: "init-b",
				PlanOutput: "plan-b",
			},
			{
				Project:        drift_stream.TypedProject{Dir: "/projects/c", Type: drift_stream.Terragrunt},
				Drifted:        false,
				Succeeded:      true,
				SkippedDueToPR: true,
			},
		},
		TotalDrifted:  1,
		TotalErrored:  &totalErrored,
		TotalSkipped:  1,
		TotalProjects: 3,
		TotalChecked:  3,
		Duration:      250 * time.Millisecond,
	}
}

func postIngest(t *testing.T, app *fiber.App, token, idemKey string, body any) (int, []byte) {
	t.Helper()
	buf, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	req := httptest.NewRequestWithContext(context.Background(),
		http.MethodPost, "/api/v1/drift_analysis", bytes.NewReader(buf))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("X-Token", token)
	}
	if idemKey != "" {
		req.Header.Set("Idempotency-Key", idemKey)
	}
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 30 * time.Second})
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, respBody
}

func TestDriftIngest_HappyPath(t *testing.T) {
	truncateAll(t)
	repoID := seedOrgAndRepo(t)
	app := newIngestApp(t)

	status, body := postIngest(t, app, seedAnalysisToken, "", sampleState())

	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", status, string(body))
	}

	var got struct {
		RunID        string `json:"run_id"`
		DashboardURL string `json:"dashboard_url"`
	}
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.RunID == "" {
		t.Fatalf("response missing run_id: %s", body)
	}
	if got.DashboardURL == "" {
		t.Fatalf("response missing dashboard_url: %s", body)
	}

	// Verify the run + 3 projects landed in the DB.
	ctx := context.Background()
	pool := withPool(t)
	var runCount, projectCount int
	if err := pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM drift_analysis_run WHERE repository_id = $1`, repoID).
		Scan(&runCount); err != nil {
		t.Fatalf("count runs: %v", err)
	}
	if runCount != 1 {
		t.Errorf("expected 1 run, got %d", runCount)
	}
	if err := pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM drift_analysis_project p
		 JOIN drift_analysis_run r ON r.uuid = p.drift_analysis_run_id
		 WHERE r.repository_id = $1`, repoID).Scan(&projectCount); err != nil {
		t.Fatalf("count projects: %v", err)
	}
	if projectCount != 3 {
		t.Errorf("expected 3 projects, got %d", projectCount)
	}

	rows, err := pool.Query(ctx,
		`SELECT p.dir, p.type, p.drifted, p.succeeded, p.init_output, p.plan_output, p.skipped_due_to_pr
		 FROM drift_analysis_project p
		 JOIN drift_analysis_run r ON r.uuid = p.drift_analysis_run_id
		 WHERE r.repository_id = $1
		 ORDER BY p.dir`, repoID)
	if err != nil {
		t.Fatalf("query projects: %v", err)
	}
	defer rows.Close()
	type projRow struct {
		dir, ptype                string
		drifted, succeeded, skipd bool
		initOut, planOut          *string
	}
	var gotRows []projRow
	for rows.Next() {
		var r projRow
		if err := rows.Scan(&r.dir, &r.ptype, &r.drifted, &r.succeeded, &r.initOut, &r.planOut, &r.skipd); err != nil {
			t.Fatalf("scan: %v", err)
		}
		gotRows = append(gotRows, r)
	}
	wantRows := []projRow{
		{dir: "/projects/a", ptype: "TERRAFORM", drifted: true, succeeded: true, initOut: ptr("init-a"), planOut: ptr("plan-a"), skipd: false},
		{dir: "/projects/b", ptype: "TOFU", drifted: false, succeeded: true, initOut: ptr("init-b"), planOut: ptr("plan-b"), skipd: false},
		{dir: "/projects/c", ptype: "TERRAGRUNT", drifted: false, succeeded: true, initOut: ptr(""), planOut: ptr(""), skipd: true},
	}
	if len(gotRows) != len(wantRows) {
		t.Fatalf("got %d rows, want %d", len(gotRows), len(wantRows))
	}
	for i := range wantRows {
		if gotRows[i].dir != wantRows[i].dir || gotRows[i].ptype != wantRows[i].ptype ||
			gotRows[i].drifted != wantRows[i].drifted || gotRows[i].succeeded != wantRows[i].succeeded ||
			gotRows[i].skipd != wantRows[i].skipd ||
			!strPtrEq(gotRows[i].initOut, wantRows[i].initOut) ||
			!strPtrEq(gotRows[i].planOut, wantRows[i].planOut) {
			t.Errorf("row %d mismatch:\n got=%+v\nwant=%+v", i, gotRows[i], wantRows[i])
		}
	}
}

func ptr(s string) *string { return &s }

func strPtrEq(a, b *string) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func TestDriftIngest_InvalidToken(t *testing.T) {
	truncateAll(t)
	seedOrgAndRepo(t)
	app := newIngestApp(t)

	status, _ := postIngest(t, app, "no-such-token", "", sampleState())
	if status != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", status)
	}
}

// TestDriftIngest_IdempotencyRace exercises the pgUniqueViolation recovery path
// in HandleUpdate: two concurrent POSTs with the same Idempotency-Key must both
// resolve to the same run UUID, and only one row should exist.
func TestDriftIngest_IdempotencyRace(t *testing.T) {
	truncateAll(t)
	repoID := seedOrgAndRepo(t)
	app := newIngestApp(t)

	const idemKey = "race-key-1"
	state := sampleState()

	var wg sync.WaitGroup
	type result struct {
		status int
		body   []byte
	}
	results := make([]result, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			status, body := postIngest(t, app, seedAnalysisToken, idemKey, state)
			results[idx] = result{status: status, body: body}
		}(i)
	}
	wg.Wait()

	for i, r := range results {
		if r.status != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d: %s", i, r.status, r.body)
		}
	}

	parse := func(b []byte) string {
		var r struct {
			RunID string `json:"run_id"`
		}
		if err := json.Unmarshal(b, &r); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		return r.RunID
	}
	runA := parse(results[0].body)
	runB := parse(results[1].body)
	if runA == "" || runB == "" {
		t.Fatalf("missing run ids: %q %q", runA, runB)
	}
	if runA != runB {
		t.Errorf("expected both concurrent calls to converge on one run, got %s vs %s", runA, runB)
	}

	var runCount int
	if err := testDB.Pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM drift_analysis_run WHERE repository_id = $1`, repoID).
		Scan(&runCount); err != nil {
		t.Fatalf("count runs: %v", err)
	}
	if runCount != 1 {
		t.Errorf("idempotent retry inserted %d rows (expected 1)", runCount)
	}
}
