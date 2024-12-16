package repository

import (
	"context"
	"driftive.cloud/api/pkg/db"
	"driftive.cloud/api/pkg/repository/queries"
)

type UserRepository interface {
	FindUserByID(ctx context.Context, id int64) (queries.User, error)
	CountUsersByProviderAndEmail(ctx context.Context, arg queries.CountUsersByProviderAndEmailParams) (int64, error)
	CreateUser(ctx context.Context, arg queries.CreateUserParams) (queries.User, error)
	FindUserByProviderAndProviderId(ctx context.Context, arg queries.FindUserByProviderAndProviderIdParams) (queries.User, error)
	FindExpiringTokensByProvider(ctx context.Context, arg queries.FindExpiringTokensByProviderParams) ([]queries.User, error)
	UpdateUserTokens(ctx context.Context, arg queries.UpdateUserTokensParams) (queries.User, error)
}

type UserRepo struct {
	db *db.DB
}

func (r *UserRepo) FindUserByID(ctx context.Context, id int64) (queries.User, error) {
	return r.db.Queries(ctx).FindUserByID(ctx, id)
}

func (r *UserRepo) CountUsersByProviderAndEmail(ctx context.Context, arg queries.CountUsersByProviderAndEmailParams) (int64, error) {
	return r.db.Queries(ctx).CountUsersByProviderAndEmail(ctx, arg)
}

func (r *UserRepo) CreateUser(ctx context.Context, arg queries.CreateUserParams) (queries.User, error) {
	return r.db.Queries(ctx).CreateUser(ctx, arg)
}

func (r *UserRepo) FindUserByProviderAndProviderId(ctx context.Context, arg queries.FindUserByProviderAndProviderIdParams) (queries.User, error) {
	return r.db.Queries(ctx).FindUserByProviderAndProviderId(ctx, arg)
}

func (r *UserRepo) FindExpiringTokensByProvider(ctx context.Context, arg queries.FindExpiringTokensByProviderParams) ([]queries.User, error) {
	return r.db.Queries(ctx).FindExpiringTokensByProvider(ctx, arg)
}

func (r *UserRepo) UpdateUserTokens(ctx context.Context, arg queries.UpdateUserTokensParams) (queries.User, error) {
	return r.db.Queries(ctx).UpdateUserTokens(ctx, arg)
}
