package repository

import (
	"context"
	"driftive.cloud/api/pkg/db"
	"driftive.cloud/api/pkg/repository/queries"
)

type GitOrgRepository interface {
	ListGitOrganizationsByProviderAndUserID(ctx context.Context, provider string, userId int64) ([]queries.GitOrganization, error)
	CreateOrUpdateGitOrganization(ctx context.Context, arg queries.CreateOrUpdateGitOrganizationParams) (queries.GitOrganization, error)
}

type GitOrgRepo struct {
	db *db.DB
}

func (g GitOrgRepo) ListGitOrganizationsByProviderAndUserID(ctx context.Context, provider string, userId int64) ([]queries.GitOrganization, error) {
	opts := queries.FindGitOrganizationByProviderAndUserIDParams{Provider: provider, UserID: userId}
	return g.db.Queries(ctx).FindGitOrganizationByProviderAndUserID(ctx, opts)
}

func (g GitOrgRepo) CreateOrUpdateGitOrganization(ctx context.Context, arg queries.CreateOrUpdateGitOrganizationParams) (queries.GitOrganization, error) {
	return g.db.Queries(ctx).CreateOrUpdateGitOrganization(ctx, arg)
}
