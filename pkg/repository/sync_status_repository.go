package repository

import (
	"context"
	"driftive.cloud/api/pkg/db"
	"driftive.cloud/api/pkg/repository/queries"
)

type SyncStatusUserRepository interface {
	CreateSyncStatusUser(ctx context.Context, arg queries.CreateSyncStatusUserParams) (queries.SyncStatusUser, error)
	FindOnePendingSyncStatusUser(ctx context.Context) (queries.SyncStatusUser, error)
}

type SyncStatusUserRepo struct {
	db *db.DB
}

func (s SyncStatusUserRepo) CreateSyncStatusUser(ctx context.Context, arg queries.CreateSyncStatusUserParams) (queries.SyncStatusUser, error) {
	return s.db.Queries(ctx).CreateSyncStatusUser(ctx, arg)
}

func (s SyncStatusUserRepo) FindOnePendingSyncStatusUser(ctx context.Context) (queries.SyncStatusUser, error) {
	return s.db.Queries(ctx).FindOnePendingSyncStatusUser(ctx)
}
