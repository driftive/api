package github

import (
	"context"
	"database/sql"
	"driftive.cloud/api/pkg/repository"
	"driftive.cloud/api/pkg/repository/queries"
	"driftive.cloud/api/pkg/usecase/utils/gh"
	"driftive.cloud/api/pkg/usecase/utils/parsing"
	"errors"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/go-github/v67/github"
	"time"
)

type SyncOrganization struct {
	orgRepository     repository.GitOrgRepository
	repoRepository    repository.GitRepositoryRepository
	orgSyncRepository repository.GitOrgSyncRepository
}

func NewSyncOrganization(orgRepository repository.GitOrgRepository, repoRepository repository.GitRepositoryRepository, gitOrgSyncRepo repository.GitOrgSyncRepository) SyncOrganization {
	return SyncOrganization{
		orgRepository:     orgRepository,
		repoRepository:    repoRepository,
		orgSyncRepository: gitOrgSyncRepo,
	}
}

func (so SyncOrganization) StartSyncLoop() {
	for {
		ctx := context.Background()
		err := so.orgSyncRepository.WithTx(ctx, func(ctx context.Context) error {
			orgSync, err := so.orgSyncRepository.FindOnePending(ctx)
			if err != nil {
				if !errors.Is(err, sql.ErrNoRows) {
					log.Errorf("error finding pending org to sync: %v", err)
				}
			}

			if orgSync.ID != 0 {
				log.Infof("syncing org: %d", orgSync.OrganizationID)
				org, err := so.orgRepository.FindGitOrgById(ctx, orgSync.OrganizationID)
				if err != nil {
					log.Errorf("error fetching org by id: %v", err)
					return err
				}
				if org.InstallationID == nil {
					so.SyncInstallationIdByOrgId(ctx, org.ID)
				}

				so.SyncOrganizationRepositories(ctx, org.ID)

				log.Infof("updating sync status for org: %d", org.ID)
				_, err = so.orgSyncRepository.UpdateSyncStatus(ctx, org.ID)
				if err != nil {
					log.Errorf("error updating sync status for user: %v", err)
				}
				log.Infof("successfully synced organization ID: %d", org.ID)
			}
			return nil
		})
		if err != nil {
			log.Errorf("error handling sync transaction: %v", err)
		}
		time.Sleep(2 * time.Second)
	}
}

func (so SyncOrganization) SyncOrganizationRepositories(ctx context.Context, orgId int64) {
	org, err := so.orgRepository.FindGitOrgById(ctx, orgId)
	if err != nil {
		log.Error("error fetching org by id: ", err)
		return
	}

	if org.InstallationID == nil {
		log.Error("no installation id found for org: ", org.Name)
		return
	}

	ghClient, err := gh.NewAppGithubInstallationClient(ctx, *org.InstallationID)
	if err != nil {
		log.Error("error creating github client: ", err)
		log.Error("aborting org sync")
		return
	}
	// paginated
	perPage := 100
	page := 1
	var allRepos []*github.Repository
	opts := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: perPage, Page: page},
	}
	for {
		repos, resp, err := ghClient.Repositories.ListByOrg(ctx, org.Name, opts)
		if err != nil {
			log.Error("error fetching repos for org: ", org.Name)
			return
		}

		allRepos = append(allRepos, repos...)

		if resp.NextPage == 0 {
			break
		}

		opts.Page = resp.NextPage
	}

	if len(allRepos) == 0 {
		log.Info("no repos found for org: ", org.Name)
		return
	}

	for _, repo := range allRepos {
		log.Info("repo: ", *repo.Name)

		params := queries.CreateOrUpdateRepositoryParams{
			OrganizationID: orgId,
			ProviderID:     parsing.Int64ToString(repo.GetID()),
			Name:           repo.GetName(),
			IsPrivate:      repo.GetPrivate(),
		}

		updatedRepo, err := so.repoRepository.CreateOrUpdateRepository(ctx, params)
		if err != nil {
			log.Error("error creating or updating repo: ", err)
			continue
		}
		log.Info("repo synced: ", updatedRepo.Name)
	}

	log.Debug("repos: ", allRepos)
}

func (so SyncOrganization) SyncInstallationIdByOrgId(ctx context.Context, orgId int64) {
	ghClient, err := gh.NewAppGithubClient(ctx)
	if err != nil {
		log.Error("error creating github client: ", err)
		log.Error("aborting org sync")
		return
	}

	org, err := so.orgRepository.FindGitOrgById(ctx, orgId)
	if err != nil {
		log.Error("error fetching org by id: ", err)
		return
	}

	installation, _, err := ghClient.Apps.FindOrganizationInstallation(ctx, org.Name)
	if err != nil {
		log.Error("error fetching installations for org: ", org.Name)
		return
	}

	if installation == nil {
		log.Error("no installation found for org: ", org.Name)
		return
	}

	err = so.orgRepository.UpdateOrgInstallationID(ctx, orgId, github.Int64(installation.GetID()))
	if err != nil {
		log.Error("error updating installation id for org: ", orgId)
		return
	}

	log.Info("installation: ", installation)
}
