package github

import (
	"context"
	"database/sql"
	"driftive.cloud/api/pkg/repository"
	"driftive.cloud/api/pkg/repository/queries"
	"driftive.cloud/api/pkg/usecase/utils/auth"
	"driftive.cloud/api/pkg/usecase/utils/gh"
	"driftive.cloud/api/pkg/usecase/utils/parsing"
	"driftive.cloud/api/pkg/usecase/utils/strutils"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/go-github/v67/github"
	"time"
)

// UserResourceSyncer syncs user resources
// Organisations and Repositories
type UserResourceSyncer struct {
	userRepository       repository.UserRepository
	gitOrgRepository     repository.GitOrgRepository
	gitRepoRepository    repository.GitRepositoryRepository
	syncStatusRepository repository.SyncStatusUserRepository
	orgSyncRepository    repository.GitOrgSyncRepository
}

func NewUserResourceSyncer(userRepo repository.UserRepository,
	gitOrgRepo repository.GitOrgRepository,
	repositoryRepository repository.GitRepositoryRepository,
	syncStatusRepository repository.SyncStatusUserRepository,
	orgSyncRepository repository.GitOrgSyncRepository) UserResourceSyncer {
	return UserResourceSyncer{
		userRepository:       userRepo,
		gitOrgRepository:     gitOrgRepo,
		gitRepoRepository:    repositoryRepository,
		syncStatusRepository: syncStatusRepository,
		orgSyncRepository:    orgSyncRepository,
	}
}

func (s *UserResourceSyncer) SyncUserResources(ctx context.Context, userId int64) error {
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
		log.Infof("Found organization: %s (Provider ID: %d)", org.GetLogin(), org.GetID())

		// Save organizations using the repository
		createOrgOpts := queries.CreateOrUpdateGitOrganizationParams{
			Provider:   "GITHUB",
			ProviderID: parsing.Int64ToString(org.GetID()),
			Name:       org.GetLogin(),
			AvatarUrl:  strutils.OrNil(org.GetAvatarURL()),
		}

		updatedOrg, err := s.gitOrgRepository.CreateOrUpdateGitOrganization(ctx, createOrgOpts)
		if err != nil {
			log.Errorf("error saving organizations for user %d: %v", userId, err)
			return err
		}

		err = s.orgSyncRepository.CreateGitOrganizationSyncIfNotExists(ctx, updatedOrg.ID)
		if err != nil {
			log.Errorf("error creating organization sync: %v", err)
			return err
		}

		userMembership, _, err := ghClient.Organizations.GetOrgMembership(ctx, user.Username, org.GetLogin())
		if err != nil {
			log.Errorf("error fetching organization membership for user %s: %v", user.Username, err)
			return err
		}

		membershipParams := queries.UpdateUserGitOrganizationMembershipParams{
			UserID:            userId,
			GitOrganizationID: updatedOrg.ID,
			Role:              gh.ParseOrgRole(*userMembership.Role),
		}
		err = s.gitOrgRepository.UpdateUserGitOrganizationMembership(ctx, membershipParams)
		if err != nil {
			log.Errorf("error updating user membership for organization: %v", err)
			return err
		}

		log.Infof("successfully saved organization: %s", updatedOrg.Name)
	}

	log.Infof("updating sync status for user: %d", userId)
	_, err = s.syncStatusRepository.UpdateSyncStatusUserLastSyncedAt(ctx, userId)
	if err != nil {
		log.Errorf("error updating sync status for user: %v", err)
	}

	log.Infof("successfully synced organizations for user: %d", userId)
	return nil
}

func (s *UserResourceSyncer) StartSyncLoop() {
	for {
		ctx := context.Background()
		err := s.syncStatusRepository.WithTx(ctx, func(ctx context.Context) error {
			result, err := s.syncStatusRepository.FindOnePendingSyncStatusUser(ctx)
			if err != nil {
				if !errors.Is(err, sql.ErrNoRows) {
					log.Errorf("error finding pending sync status user: %v", err)
				}
			}

			if result.ID != 0 {
				err = s.SyncUserResources(ctx, result.UserID)
				if err != nil {
					log.Errorf("error syncing user resources: %v", err)
				}
			} else {
				log.Debug("no pending sync status user found")
				time.Sleep(5 * time.Second)
			}
			return nil
		})
		if err != nil {
			log.Errorf("error handling sync transaction: %v", err)
		}
	}
}

func (s *UserResourceSyncer) HandleUserSyncRequest(c *fiber.Ctx) error {
	userId, err := auth.MustGetLoggedUserId(c)
	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	log.Infof("syncing organizations for user: %d", userId)

	err = s.SyncUserResources(c.Context(), *userId)
	if err != nil {
		log.Errorf("error syncing organizations for user: %d: %v", userId, err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	return c.SendStatus(fiber.StatusOK)
}
