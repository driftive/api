package repository

import (
	"context"
	"driftive.cloud/api/pkg/db"
	"driftive.cloud/api/pkg/repository/queries"
)

type GitRepositoryRepository interface {
	FindGitRepositoryByID(ctx context.Context, id int64) (queries.GitRepository, error)
	CreateOrUpdateRepository(ctx context.Context, orgId int64, providerId string, name string) (queries.GitRepository, error)
	//FindGitRepositoryByProviderAndProviderId(ctx context.Context, arg queries.FindGitRepositoryByProviderAndProviderIdParams) (queries.GitRepository, error)
	//FindGitRepositoryByProviderAndOwnerAndName(ctx context.Context, arg queries.FindGitRepositoryByProviderAndOwnerAndNameParams) (queries.GitRepository, error)
}

type GitRepoRepo struct {
	db *db.DB
}

func (r *GitRepoRepo) FindGitRepositoryByID(ctx context.Context, id int64) (queries.GitRepository, error) {
	return r.db.Queries(ctx).FindGitRepositoryById(ctx, id)
}

func (r *GitRepoRepo) CreateOrUpdateRepository(ctx context.Context, orgId int64, providerId, name string) (queries.GitRepository, error) {
	params := queries.CreateOrUpdateRepositoryParams{
		OrganizationID: orgId,
		ProviderID:     providerId,
		Name:           name,
	}
	return r.db.Queries(ctx).CreateOrUpdateRepository(ctx, params)
}

//func (r *GitRepositoryRepo) FindGitRepositoryByProviderAndProviderId(ctx context.Context, arg queries.FindGitRepositoryByProviderAndProviderIdParams) (queries.GitRepository, error) {
//	return r.db.Queries(ctx).FindGitRepositoryByProviderAndProviderId(ctx, arg)
//}
//
//func (r *GitRepositoryRepo) FindGitRepositoryByProviderAndOwnerAndName(ctx context.Context, arg queries.FindGitRepositoryByProviderAndOwnerAndNameParams) (queries.GitRepository, error) {
//	return r.db.Queries(ctx).FindGitRepositoryByProviderAndOwnerAndName(ctx, arg)
//}
