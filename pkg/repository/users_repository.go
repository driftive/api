package repository

import (
	"context"
	"driftive.cloud/api/pkg/db"
	"driftive.cloud/api/pkg/repository/queries"
)

type UserRepository interface {
	FindUserByID(ctx context.Context, id int64) (queries.User, error)
	CountUsersByProviderAndProviderId(ctx context.Context, arg queries.CountUsersByProviderAndProviderIdParams) (int64, error)
	CreateOrUpdateUser(ctx context.Context, arg queries.CreateOrUpdateUserParams) (queries.User, error)
	FindUserByProviderAndProviderId(ctx context.Context, arg queries.FindUserByProviderAndProviderIdParams) (queries.User, error)
	FindExpiringTokensByProvider(ctx context.Context, arg queries.FindExpiringTokensByProviderParams) ([]queries.User, error)
	FindAndLockExpiringToken(ctx context.Context, arg queries.FindAndLockExpiringTokenParams) (queries.User, error)
	UpdateUserTokens(ctx context.Context, arg queries.UpdateUserTokensParams) (queries.User, error)
	IncrementTokenRefreshAttempts(ctx context.Context, id int64) (queries.User, error)
	DisableTokenRefresh(ctx context.Context, id int64) (queries.User, error)
	WithTx(ctx context.Context, fn func(context.Context) error) error
}

type UserRepo struct {
	db *db.DB
}

func (r *UserRepo) FindUserByID(ctx context.Context, id int64) (queries.User, error) {
	return r.db.Queries(ctx).FindUserByID(ctx, id)
}

func (r *UserRepo) CountUsersByProviderAndProviderId(ctx context.Context, arg queries.CountUsersByProviderAndProviderIdParams) (int64, error) {
	return r.db.Queries(ctx).CountUsersByProviderAndProviderId(ctx, arg)
}

func (r *UserRepo) CreateOrUpdateUser(ctx context.Context, arg queries.CreateOrUpdateUserParams) (queries.User, error) {
	return r.db.Queries(ctx).CreateOrUpdateUser(ctx, arg)
}

func (r *UserRepo) FindUserByProviderAndProviderId(ctx context.Context, arg queries.FindUserByProviderAndProviderIdParams) (queries.User, error) {
	return r.db.Queries(ctx).FindUserByProviderAndProviderId(ctx, arg)
}

func (r *UserRepo) FindExpiringTokensByProvider(ctx context.Context, arg queries.FindExpiringTokensByProviderParams) ([]queries.User, error) {
	return r.db.Queries(ctx).FindExpiringTokensByProvider(ctx, arg)
}

func (r *UserRepo) FindAndLockExpiringToken(ctx context.Context, arg queries.FindAndLockExpiringTokenParams) (queries.User, error) {
	return r.db.Queries(ctx).FindAndLockExpiringToken(ctx, arg)
}

func (r *UserRepo) UpdateUserTokens(ctx context.Context, arg queries.UpdateUserTokensParams) (queries.User, error) {
	return r.db.Queries(ctx).UpdateUserTokens(ctx, arg)
}

func (r *UserRepo) IncrementTokenRefreshAttempts(ctx context.Context, id int64) (queries.User, error) {
	return r.db.Queries(ctx).IncrementTokenRefreshAttempts(ctx, id)
}

func (r *UserRepo) DisableTokenRefresh(ctx context.Context, id int64) (queries.User, error) {
	return r.db.Queries(ctx).DisableTokenRefresh(ctx, id)
}

func (r *UserRepo) WithTx(ctx context.Context, fn func(context.Context) error) error {
	return r.db.WithTx(ctx, fn)
}
