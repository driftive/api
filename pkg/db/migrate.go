package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"strconv"

	"driftive.cloud/api/pkg/config"
	"driftive.cloud/api/pkg/utils"

	"github.com/gofiber/fiber/v3/log"
	"github.com/golang-migrate/migrate/v4"
	migratepgx "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func RunMigrations(ctx context.Context, cfg config.Config, fsys fs.FS) error {
	if !boolEnv("AUTO_MIGRATE", true) {
		log.Info("AUTO_MIGRATE disabled, skipping migrations")
		return nil
	}

	hostPort := net.JoinHostPort(cfg.Database.Host, strconv.Itoa(cfg.Database.Port))
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?",
		cfg.Database.User, cfg.Database.Password, hostPort, cfg.Database.Database)

	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open migration db: %w", err)
	}
	defer sqlDB.Close()

	if err := baselineFromFlyway(ctx, sqlDB); err != nil {
		return fmt.Errorf("flyway baseline: %w", err)
	}

	src, err := iofs.New(fsys, "migrations")
	if err != nil {
		return fmt.Errorf("open migration source: %w", err)
	}

	driver, err := migratepgx.WithInstance(sqlDB, &migratepgx.Config{})
	if err != nil {
		return fmt.Errorf("init migration driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, "pgx5", driver)
	if err != nil {
		return fmt.Errorf("init migrator: %w", err)
	}

	log.Info("running database migrations...")
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}
	log.Info("database migrations up to date")
	return nil
}

func boolEnv(key string, def bool) bool {
	parsed, err := strconv.ParseBool(utils.GetEnvOrDefault(key, strconv.FormatBool(def)))
	if err != nil {
		return def
	}
	return parsed
}
