package repository

import (
	"context"
	"driftive.cloud/api/pkg/config"
	"driftive.cloud/api/pkg/db"
	"driftive.cloud/api/pkg/repository/queries"
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

func (r *Repository) UserRepository() UserRepository {
	return &UserRepo{db: r.db}
}
