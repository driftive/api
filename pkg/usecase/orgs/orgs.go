package orgs

import (
	"driftive.cloud/api/pkg/config"
	"driftive.cloud/api/pkg/db"
	"driftive.cloud/api/pkg/repository"
	"driftive.cloud/api/pkg/usecase/utils/fiberutils"
	"driftive.cloud/api/pkg/usecase/utils/parsing"
	"github.com/gofiber/fiber/v2"
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
	userId := fiberutils.GetUserId(c)
	orgs, err := h.gitOrgRepository.ListGitOrganizationsByProviderAndUserID(c.Context(), "GITHUB", userId)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	return c.JSON(parsing.ToOrganizationDTOs(orgs))
}
