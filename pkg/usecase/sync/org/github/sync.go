package github

import (
	"context"
	"driftive.cloud/api/pkg/repository"
	"driftive.cloud/api/pkg/usecase/utils/gh"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/go-github/v67/github"
)

type SyncOrganization struct {
	orgRepository repository.GitOrgRepository
}

func NewSyncOrganization(orgRepository repository.GitOrgRepository) SyncOrganization {
	return SyncOrganization{
		orgRepository: orgRepository,
	}
}

func (so SyncOrganization) SyncById(orgId int64) {
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
