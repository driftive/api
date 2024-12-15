package repository

import (
	"context"
	"driftive.cloud/api/pkg/db"
	"driftive.cloud/api/pkg/repository/queries"
)

type UserRepository interface {
	CountUsersByProviderAndEmail(ctx context.Context, arg queries.CountUsersByProviderAndEmailParams) (int64, error)
	CreateUser(ctx context.Context, arg queries.CreateUserParams) (queries.User, error)
	FindUserByProviderAndProviderId(ctx context.Context, arg queries.FindUserByProviderAndProviderIdParams) (queries.User, error)
}

type UserRepo struct {
	db *db.DB
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
