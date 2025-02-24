package github

import (
	"context"
	"driftive.cloud/api/pkg/repository"
	"driftive.cloud/api/pkg/usecase/utils/gh"
	"driftive.cloud/api/pkg/usecase/utils/parsing"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/go-github/v67/github"
)

type SyncOrganization struct {
	orgRepository  repository.GitOrgRepository
	repoRepository repository.GitRepositoryRepository
}

func NewSyncOrganization(orgRepository repository.GitOrgRepository, repoRepository repository.GitRepositoryRepository) SyncOrganization {
	return SyncOrganization{
		orgRepository:  orgRepository,
		repoRepository: repoRepository,
	}
}

func (so SyncOrganization) SyncOrganizationRepositories(orgId int64) {
	ctx := context.Background()

	org, err := so.orgRepository.FindGitOrgById(ctx, orgId)
	if err != nil {
		log.Error("error fetching org by id: ", err)
		return
	}

	if org.InstallationID == nil {
		log.Error("no installation id found for org: ", org.Name)
		return
	}

	ghClient, err := gh.NewAppGithubInstallationClient(*org.InstallationID)
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
		log.Info("repo: ", repo.Name)
		repo, err := so.repoRepository.CreateOrUpdateRepository(ctx, orgId, parsing.Int64ToString(repo.GetID()), repo.GetName())
		if err != nil {
			log.Error("error creating or updating repo: ", err)
			continue
		}
		log.Info("repo synced: ", repo.Name)
	}

	log.Info("repos: ", allRepos)
}

func (so SyncOrganization) SyncInstallationIdByOrgId(orgId int64) {
	ctx := context.Background()

	ghClient, err := gh.NewAppGithubClient()
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
