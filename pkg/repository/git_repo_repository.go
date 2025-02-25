package repository

import (
	"context"
	"driftive.cloud/api/pkg/db"
	"driftive.cloud/api/pkg/repository/queries"
)

type GitRepositoryRepository interface {
	FindGitRepositoryById(ctx context.Context, id int64) (queries.GitRepository, error)
	CreateOrUpdateRepository(ctx context.Context, params queries.CreateOrUpdateRepositoryParams) (queries.GitRepository, error)
	FindGitReposByOrgId(ctx context.Context, orgId int64) ([]queries.GitRepository, error)
	FindGitRepositoryByOrgIdAndName(ctx context.Context, orgId int64, repoName string) (queries.GitRepository, error)
	UpdateRepositoryToken(ctx context.Context, params queries.UpdateRepositoryTokenParams) (*string, error)
	FindGitRepositoryByToken(ctx context.Context, token string) (queries.GitRepository, error)
}

type GitRepoRepo struct {
	db *db.DB
}

func (r *GitRepoRepo) FindGitRepositoryById(ctx context.Context, id int64) (queries.GitRepository, error) {
	return r.db.Queries(ctx).FindGitRepositoryById(ctx, id)
}

func (r *GitRepoRepo) CreateOrUpdateRepository(ctx context.Context, params queries.CreateOrUpdateRepositoryParams) (queries.GitRepository, error) {
	return r.db.Queries(ctx).CreateOrUpdateRepository(ctx, params)
}

func (r *GitRepoRepo) FindGitReposByOrgId(ctx context.Context, orgId int64) ([]queries.GitRepository, error) {
	return r.db.Queries(ctx).FindGitRepositoriesByOrgId(ctx, orgId)
}

func (r *GitRepoRepo) FindGitRepositoryByOrgIdAndName(ctx context.Context, orgId int64, repoName string) (queries.GitRepository, error) {
	params := queries.FindGitRepositoryByOrgIdAndNameParams{
		OrganizationID: orgId,
		Name:           repoName,
	}
	return r.db.Queries(ctx).FindGitRepositoryByOrgIdAndName(ctx, params)
}

func (r *GitRepoRepo) UpdateRepositoryToken(ctx context.Context, params queries.UpdateRepositoryTokenParams) (*string, error) {
	return r.db.Queries(ctx).UpdateRepositoryToken(ctx, params)
}

func (r *GitRepoRepo) FindGitRepositoryByToken(ctx context.Context, token string) (queries.GitRepository, error) {
	return r.db.Queries(ctx).FindGitRepositoryByToken(ctx, &token)
}
