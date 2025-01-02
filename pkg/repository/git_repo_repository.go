package repository

import (
	"context"
	"driftive.cloud/api/pkg/db"
	"driftive.cloud/api/pkg/repository/queries"
)

type GitRepositoryRepository interface {
	FindGitRepositoryById(ctx context.Context, id int64) (queries.GitRepository, error)
	//FindGitRepositoryByProviderAndProviderId(ctx context.Context, arg queries.FindGitRepositoryByProviderAndProviderIdParams) (queries.GitRepository, error)
	//FindGitRepositoryByProviderAndOwnerAndName(ctx context.Context, arg queries.FindGitRepositoryByProviderAndOwnerAndNameParams) (queries.GitRepository, error)
}

type GitRepositoryRepo struct {
	db *db.DB
}

func (r *GitRepositoryRepo) FindGitRepositoryByID(ctx context.Context, id int64) (queries.GitRepository, error) {
	return r.db.Queries(ctx).FindGitRepositoryById(ctx, id)
}

//func (r *GitRepositoryRepo) FindGitRepositoryByProviderAndProviderId(ctx context.Context, arg queries.FindGitRepositoryByProviderAndProviderIdParams) (queries.GitRepository, error) {
//	return r.db.Queries(ctx).FindGitRepositoryByProviderAndProviderId(ctx, arg)
//}
//
//func (r *GitRepositoryRepo) FindGitRepositoryByProviderAndOwnerAndName(ctx context.Context, arg queries.FindGitRepositoryByProviderAndOwnerAndNameParams) (queries.GitRepository, error) {
//	return r.db.Queries(ctx).FindGitRepositoryByProviderAndOwnerAndName(ctx, arg)
//}
