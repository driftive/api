package repository

import (
	"context"
	"driftive.cloud/api/pkg/db"
	"driftive.cloud/api/pkg/repository/queries"
)

type GitOrgRepository interface {
	ListGitOrganizationsByProviderAndUserID(ctx context.Context, provider string, userId int64) ([]queries.GitOrganization, error)
	CreateOrUpdateGitOrganization(ctx context.Context, arg queries.CreateOrUpdateGitOrganizationParams) (queries.GitOrganization, error)
	UpdateUserGitOrganizationMembership(ctx context.Context, arg queries.UpdateUserGitOrganizationMembershipParams) error
	FindGitOrgById(ctx context.Context, id int64) (queries.GitOrganization, error)
	UpdateOrgInstallationID(ctx context.Context, orgId int64, installationId *int64) error
}

type GitOrgRepo struct {
	db *db.DB
}

func (g GitOrgRepo) ListGitOrganizationsByProviderAndUserID(ctx context.Context, provider string, userId int64) ([]queries.GitOrganization, error) {
	opts := queries.FindGitOrganizationByProviderAndUserIDParams{Provider: provider, UserID: userId}
	orgs, err := g.db.Queries(ctx).FindGitOrganizationByProviderAndUserID(ctx, opts)
	if err != nil {
		return nil, err
	}
	if orgs == nil {
		return []queries.GitOrganization{}, nil
	}
	return orgs, nil
}

func (g GitOrgRepo) CreateOrUpdateGitOrganization(ctx context.Context, arg queries.CreateOrUpdateGitOrganizationParams) (queries.GitOrganization, error) {
	return g.db.Queries(ctx).CreateOrUpdateGitOrganization(ctx, arg)
}

func (g GitOrgRepo) UpdateUserGitOrganizationMembership(ctx context.Context, arg queries.UpdateUserGitOrganizationMembershipParams) error {
	return g.db.Queries(ctx).UpdateUserGitOrganizationMembership(ctx, arg)
}

func (g GitOrgRepo) FindGitOrgById(ctx context.Context, id int64) (queries.GitOrganization, error) {
	return g.db.Queries(ctx).FindGitOrganizationByID(ctx, id)
}

func (g GitOrgRepo) UpdateOrgInstallationID(ctx context.Context, orgId int64, installationId *int64) error {
	params := queries.UpdateOrgInstallationIDParams{ID: orgId, InstallationID: installationId}
	return g.db.Queries(ctx).UpdateOrgInstallationID(ctx, params)
}
