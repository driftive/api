package repository

import (
	"context"
	"driftive.cloud/api/pkg/db"
	"driftive.cloud/api/pkg/repository/queries"
)

type SyncStatusUserRepository interface {
	WithTx(ctx context.Context, fn func(context.Context) error) error
	CreateOrUpdateSyncStatusUser(ctx context.Context, userID int64) (queries.SyncStatusUser, error)
	FindOnePendingSyncStatusUser(ctx context.Context) (queries.SyncStatusUser, error)
	UpdateSyncStatusUserLastSyncedAt(ctx context.Context, syncStatusUserID int64) (queries.SyncStatusUser, error)
	FindSyncStatusUserByUserID(ctx context.Context, userID int64) (queries.SyncStatusUser, error)
}

type SyncStatusUserRepo struct {
	db *db.DB
}

func (s SyncStatusUserRepo) CreateOrUpdateSyncStatusUser(ctx context.Context, userID int64) (queries.SyncStatusUser, error) {
	return s.db.Queries(ctx).CreateOrUpdateSyncStatusUser(ctx, userID)
}

func (s SyncStatusUserRepo) FindOnePendingSyncStatusUser(ctx context.Context) (queries.SyncStatusUser, error) {
	return s.db.Queries(ctx).FindOnePendingSyncStatusUser(ctx)
}

func (s SyncStatusUserRepo) UpdateSyncStatusUserLastSyncedAt(ctx context.Context, syncStatusUserID int64) (queries.SyncStatusUser, error) {
	return s.db.Queries(ctx).UpdateSyncStatusUserLastSyncedAt(ctx, syncStatusUserID)
}

func (s SyncStatusUserRepo) FindSyncStatusUserByUserID(ctx context.Context, userID int64) (queries.SyncStatusUser, error) {
	return s.db.Queries(ctx).FindSyncStatusUserByUserID(ctx, userID)
}

func (s SyncStatusUserRepo) WithTx(ctx context.Context, fn func(context.Context) error) error {
	return s.db.WithTx(ctx, fn)
}
