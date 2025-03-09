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
	userId, err := auth.MustGetLoggedUserId(c)
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
		err = auth.MustHavePermission(c, org.ID)
		if err != nil {
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		return c.JSON(parsing.ToOrganizationDTO(org))
	default:
		break
	}
	return c.SendStatus(fiber.StatusNotImplemented)
}

// HandleGHOrganizationInstalled handles the installation of a GitHub organization
// Currently, it just logs the installation ID and redirects to the frontend
// TODO: Implement the actual handling of the installation
func (h *GitOrganizationHandler) HandleGHOrganizationInstalled(c *fiber.Ctx) error {
	log.Infof("GH organization installed. Installation ID: %s. Setup action: %s",
		c.Query("installation_id"), c.Query("setup_action"))
	return c.Redirect(h.cfg.Frontend.FrontendURL + "/gh/orgs")
}
