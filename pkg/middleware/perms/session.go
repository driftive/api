package perms

import (
	"driftive.cloud/api/pkg/repository"
	"driftive.cloud/api/pkg/usecase/utils/auth"
	"github.com/gofiber/fiber/v2"
)

func New(orgRepo repository.GitOrgRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userId, err := auth.MustGetLoggedUserId(c)
		if err == nil {
			orgIds, err := orgRepo.FindAllUserOrganizationIds(c.Context(), *userId)
			if err != nil {
				return c.SendStatus(fiber.StatusInternalServerError)
			}
			c.Locals("user_org_ids", orgIds)
		}
		return c.Next()
	}
}
