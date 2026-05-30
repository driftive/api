// Package integration brings up a real Postgres via dockertest, applies the
// repo's Flyway migrations as plain SQL, and exposes a shared *db.DB to each
// test. Requires Docker. Skipped if DOCKER_HOST / docker socket is unreachable.
//
// Run with:  go test ./test/integration/...
package integration

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"driftive.cloud/api/pkg/config"
	"driftive.cloud/api/pkg/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/moby/moby/api/types/container"
	"github.com/ory/dockertest/v4"
)

// testDB is the *db.DB shared by every test. Populated by TestMain.
var testDB *db.DB

// migrationsDir resolves to api/migrations/ regardless of where the test is run from.
var migrationsDir string

const (
	pgUser     = "driftive"
	pgPassword = "driftive"
	pgDatabase = "driftive"
)

func TestMain(m *testing.M) {
	// Best-effort discovery of the migrations directory so the test runs from
	// either the package dir or the repo root.
	for _, candidate := range []string{"../../migrations", "migrations"} {
		if abs, err := filepath.Abs(candidate); err == nil {
			if info, err := os.Stat(abs); err == nil && info.IsDir() {
				migrationsDir = abs
				break
			}
		}
	}
	if migrationsDir == "" {
		log.Fatal("integration tests: could not locate migrations directory")
	}

	ctx := context.Background()
	pool, err := dockertest.NewPool(ctx, "", dockertest.WithMaxWait(60*time.Second))
	if err != nil {
		log.Printf("integration tests: docker unreachable, skipping: %v", err)
		os.Exit(0)
	}

	resource, err := pool.Run(ctx, "postgres",
		dockertest.WithTag("16-alpine"),
		dockertest.WithEnv([]string{
			"POSTGRES_USER=" + pgUser,
			"POSTGRES_PASSWORD=" + pgPassword,
			"POSTGRES_DB=" + pgDatabase,
			"listen_addresses=*",
		}),
		dockertest.WithHostConfig(func(hc *container.HostConfig) {
			hc.AutoRemove = true
			hc.RestartPolicy = container.RestartPolicy{Name: container.RestartPolicyDisabled}
		}),
	)
	if err != nil {
		log.Fatalf("integration tests: failed to start postgres: %v", err)
	}

	hostPort := resource.GetHostPort("5432/tcp")
	host, portStr, err := net.SplitHostPort(hostPort)
	if err != nil {
		log.Fatalf("integration tests: invalid host:port %q: %v", hostPort, err)
	}
	port, _ := strconv.Atoi(portStr)

	connStr := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		pgUser, pgPassword, net.JoinHostPort(host, portStr), pgDatabase)
	if err := pool.Retry(ctx, 60*time.Second, func() error {
		conn, err := pgx.Connect(ctx, connStr)
		if err != nil {
			return err
		}
		defer conn.Close(ctx)
		return conn.Ping(ctx)
	}); err != nil {
		log.Fatalf("integration tests: postgres never became reachable: %v", err)
	}

	if err := applyMigrations(connStr); err != nil {
		log.Fatalf("integration tests: migrations failed: %v", err)
	}

	cfg := config.Config{
		Database: config.Database{
			User:        pgUser,
			Password:    pgPassword,
			Host:        host,
			Port:        port,
			Database:    pgDatabase,
			Connections: 4,
		},
	}
	testDB = db.NewDB(cfg)

	code := m.Run()

	testDB.Pool.Close()
	if err := pool.Close(ctx); err != nil {
		log.Printf("integration tests: failed to close docker pool: %v", err)
	}
	os.Exit(code)
}

// applyMigrations runs every *.sql file under migrationsDir in filename order.
// The Flyway naming convention (V<timestamp>__name.sql) sorts lexicographically
// to the right order, so plain sort.Strings is enough.
func applyMigrations(connStr string) error {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return fmt.Errorf("connect for migrations: %w", err)
	}
	defer conn.Close(ctx)

	for _, name := range names {
		body, err := os.ReadFile(filepath.Join(migrationsDir, name))
		if err != nil {
			return fmt.Errorf("read %s: %w", name, err)
		}
		if _, err := conn.Exec(ctx, string(body)); err != nil {
			return fmt.Errorf("apply %s: %w", name, err)
		}
	}
	return nil
}

// truncateAll wipes test data between cases. Uses TRUNCATE ... CASCADE so it's
// safe regardless of FK order. Skip if testDB hasn't been initialised (Docker
// unreachable; TestMain returned 0 above).
func truncateAll(t *testing.T) {
	t.Helper()
	if testDB == nil {
		t.Skip("integration tests skipped (no testDB)")
	}
	tables := []string{
		"drift_analysis_project",
		"drift_analysis_run",
		"git_repository",
		"user_git_organization",
		"git_organization",
		"sync_status_user",
		"git_organization_sync",
		"users",
	}
	for _, tbl := range tables {
		_, err := testDB.Pool.Exec(context.Background(), "TRUNCATE TABLE "+tbl+" RESTART IDENTITY CASCADE")
		if err != nil {
			// Table may not exist on all schemas — log but don't fail.
			t.Logf("truncate %s: %v", tbl, err)
		}
	}
}

// withPool returns the raw pgxpool for direct seeding helpers.
func withPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	if testDB == nil {
		t.Skip("integration tests skipped (no testDB)")
	}
	return testDB.Pool
}
