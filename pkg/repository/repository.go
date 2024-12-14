package repository

import (
	"context"
	"driftive.cloud/api/pkg/config"
	"driftive.cloud/api/pkg/db"
	"driftive.cloud/api/pkg/repository/queries"
	"github.com/jackc/pgx/v5"
)

type Repositories interface {
	Queries(ctx context.Context) *queries.Queries
}

type Repository struct {
	db     *db.DB
	config *config.Config
}

func NewRepository(db *db.DB, config *config.Config) Repository {
	return Repository{db: db, config: config}
}

func (r *Repository) Queries(ctx context.Context) *queries.Queries {
	if ctx.Value("tx") != nil {
		return queries.New(r.db.Pool).WithTx(ctx.Value("tx").(pgx.Tx))
	}
	return queries.New(r.db.Pool)
}
