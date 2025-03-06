package orgs

import (
	"driftive.cloud/api/pkg/config"
	"driftive.cloud/api/pkg/db"
	"driftive.cloud/api/pkg/model"
	"driftive.cloud/api/pkg/repository"
	"driftive.cloud/api/pkg/usecase/utils/auth"
	"driftive.cloud/api/pkg/usecase/utils/parsing"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

type GitOrganizationHandler struct {
	cfg              config.Config
	db               *db.DB
	gitOrgRepository repository.GitOrgRepository
}

func NewGitOrganizationHandler(cfg config.Config, db *db.DB, orgRepo repository.GitOrgRepository) *GitOrganizationHandler {
	return &GitOrganizationHandler{
		cfg:              cfg,
		db:               db,
		gitOrgRepository: orgRepo,
	}
}

func (h *GitOrganizationHandler) ListGitOrganizations(c *fiber.Ctx) error {
	userId, err := auth.GetLoggedUserId(c)
	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	log.Infof("fetching organizations for user: %d", userId)
	orgs, err := h.gitOrgRepository.ListGitOrganizationsByProviderAndUserID(c.Context(), "GITHUB", *userId)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	return c.JSON(parsing.ToOrganizationDTOs(orgs))
}

func (h *GitOrganizationHandler) GetOrgByNameAndProvider(c *fiber.Ctx, provider model.GitProvider) error {
	userId, err := auth.GetLoggedUserId(c)
	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	orgName := c.Query("org_name")
	if orgName == "" {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	switch provider {
	case model.GitHubProvider:
		org, err := h.gitOrgRepository.FindGitOrgByProviderAndName(c.Context(), "GITHUB", orgName)
		if err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		// Check if user is a member of the organization
		isMember, err := h.gitOrgRepository.IsUserMemberOfOrg(c.Context(), org.ID, *userId)
		if err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		if !isMember {
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		return c.JSON(parsing.ToOrganizationDTO(org))
	default:
		break
	}
	return c.SendStatus(fiber.StatusNotImplemented)
}
