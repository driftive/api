package repository

import (
	"context"
	"driftive.cloud/api/pkg/db"
	"driftive.cloud/api/pkg/repository/queries"
)

type SyncStatusUserRepository interface {
	CreateSyncStatusUser(ctx context.Context, userID int64) (queries.SyncStatusUser, error)
	FindOnePendingSyncStatusUser(ctx context.Context) (queries.SyncStatusUser, error)
	UpdateSyncStatusUserLastSyncedAt(ctx context.Context, syncStatusUserID int64) (queries.SyncStatusUser, error)
}

type SyncStatusUserRepo struct {
	db *db.DB
}

func (s SyncStatusUserRepo) CreateSyncStatusUser(ctx context.Context, userID int64) (queries.SyncStatusUser, error) {
	return s.db.Queries(ctx).CreateSyncStatusUser(ctx, userID)
}

func (s SyncStatusUserRepo) FindOnePendingSyncStatusUser(ctx context.Context) (queries.SyncStatusUser, error) {
	return s.db.Queries(ctx).FindOnePendingSyncStatusUser(ctx)
}

func (s SyncStatusUserRepo) UpdateSyncStatusUserLastSyncedAt(ctx context.Context, syncStatusUserID int64) (queries.SyncStatusUser, error) {
	return s.db.Queries(ctx).UpdateSyncStatusUserLastSyncedAt(ctx, syncStatusUserID)
}
