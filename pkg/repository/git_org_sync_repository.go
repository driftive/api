package repository

import (
	"context"
	"driftive.cloud/api/pkg/db"
	"driftive.cloud/api/pkg/repository/queries"
)

type GitOrgSyncRepository interface {
	CreateGitOrganizationSyncIfNotExists(ctx context.Context, orgId int64) error
	FindOnePending(ctx context.Context) (queries.GitOrganizationSync, error)
	UpdateSyncStatus(ctx context.Context, orgId int64) (queries.GitOrganizationSync, error)
	WithTx(ctx context.Context, fn func(context.Context) error) error
}

type GitOrgSyncRepo struct {
	db *db.DB
}

func (g GitOrgSyncRepo) CreateGitOrganizationSyncIfNotExists(ctx context.Context, orgId int64) error {
	return g.db.Queries(ctx).CreateGitOrganizationSyncIfNotExists(ctx, orgId)
}

func (g GitOrgSyncRepo) FindOnePending(ctx context.Context) (queries.GitOrganizationSync, error) {
	return g.db.Queries(ctx).FindOnePendingSyncOrg(ctx)
}

func (g GitOrgSyncRepo) UpdateSyncStatus(ctx context.Context, orgId int64) (queries.GitOrganizationSync, error) {
	return g.db.Queries(ctx).UpdateGitOrganizationSyncStatus(ctx, orgId)
}

func (g GitOrgSyncRepo) WithTx(ctx context.Context, fn func(context.Context) error) error {
	return g.db.WithTx(ctx, fn)
}
