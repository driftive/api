package db

// Transitional: delete this file and its call in RunMigrations once every
// environment has booted the golang-migrate binary at least once.

import (
	"context"
	"database/sql"

	"github.com/gofiber/fiber/v3/log"
)

func baselineFromFlyway(ctx context.Context, sqlDB *sql.DB) error {
	var schemaMigrations, flywayHistory *string
	err := sqlDB.QueryRowContext(ctx,
		`SELECT to_regclass('public.schema_migrations')::text, to_regclass('public.flyway_schema_history')::text`,
	).Scan(&schemaMigrations, &flywayHistory)
	if err != nil {
		return err
	}
	if schemaMigrations != nil || flywayHistory == nil {
		return nil
	}

	var maxVersion *int64
	if err := sqlDB.QueryRowContext(ctx,
		`SELECT MAX(version::bigint) FROM flyway_schema_history WHERE success`,
	).Scan(&maxVersion); err != nil {
		return err
	}
	if maxVersion == nil {
		return nil
	}

	if _, err := sqlDB.ExecContext(ctx,
		`CREATE TABLE schema_migrations (version bigint NOT NULL PRIMARY KEY, dirty boolean NOT NULL)`,
	); err != nil {
		return err
	}
	if _, err := sqlDB.ExecContext(ctx,
		`INSERT INTO schema_migrations (version, dirty) VALUES ($1, false)`, *maxVersion,
	); err != nil {
		return err
	}
	log.Infof("baselined schema_migrations from flyway_schema_history at version %d", *maxVersion)
	return nil
}
