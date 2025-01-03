package github

import (
	"context"
	"driftive.cloud/api/pkg/repository"
	"driftive.cloud/api/pkg/repository/queries"
	"driftive.cloud/api/pkg/usecase/utils/gh"
	"driftive.cloud/api/pkg/usecase/utils/strutils"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/go-github/v67/github"
	"time"
)

// UserResourceSyncer syncs user resources
// Organisations and Repositories
type UserResourceSyncer struct {
	userRepository    repository.UserRepository
	gitOrgRepository  repository.GitOrgRepository
	gitRepoRepository repository.GitRepositoryRepository
}

func NewUserResourceSyncer(userRepo repository.UserRepository, gitOrgRepo repository.GitOrgRepository, repositoryRepository repository.GitRepositoryRepository) UserResourceSyncer {
	return UserResourceSyncer{
		userRepository:    userRepo,
		gitOrgRepository:  gitOrgRepo,
		gitRepoRepository: repositoryRepository,
	}
}

func (s *UserResourceSyncer) SyncUserResources(userId int64) error {
	ctx := context.Background()

	log.Info("syncing user resources for user: ", userId)

	user, err := s.userRepository.FindUserByID(ctx, userId)
	if err != nil {
		log.Errorf("error finding user by id: %v", err)
	}

	ghClient := gh.NewDefaultGithubClient(user.AccessToken)

	var allOrgs []*github.Organization
	opts := &github.ListOptions{PerPage: 100} // Fetch up to 100 orgs per page

	// Loop to fetch paginated results
	for {
		orgs, resp, err := ghClient.Organizations.List(ctx, "", opts)
		if err != nil {
			log.Errorf("error fetching organizations for user %d: %v", user.ID, err)
			return err
		}

		// Append the fetched organizations to the result
		allOrgs = append(allOrgs, orgs...)

		// Check if there are more pages
		if resp.NextPage == 0 {
			break // Exit the loop when no more pages are available
		}

		// Move to the next page
		opts.Page = resp.NextPage
	}

	// Print or process the organizations
	for _, org := range allOrgs {
		log.Infof("Found organization: %s (ID: %d)", org.GetLogin(), org.GetID())

		// Save organizations using the repository
		createOrgOpts := queries.CreateOrUpdateGitOrganizationParams{
			Provider:   "GITHUB",
			ProviderID: strutils.Int64ToString(org.GetID()),
			Name:       org.GetName(),
		}

		updatedOrg, err := s.gitOrgRepository.CreateOrUpdateGitOrganization(ctx, createOrgOpts)
		if err != nil {
			log.Errorf("error saving organizations for user %d: %v", userId, err)
			return err
		}

		// update membership

		log.Infof("successfully saved organization: %s", updatedOrg.Name)
	}

	log.Infof("successfully synced organizations for user: %d", userId)
	return nil
}

func (s *UserResourceSyncer) StartSyncLoop() {
	for {

		time.Sleep(10 * time.Hour)
	}
}
