package auth

import (
	"driftive.cloud/api/pkg/repository"
	"driftive.cloud/api/pkg/usecase/utils/auth"
	"driftive.cloud/api/pkg/usecase/utils/gh"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

type ProfileHandler struct {
	userRepo repository.UserRepository
}

func NewProfileHandler(userRepo repository.UserRepository) ProfileHandler {
	return ProfileHandler{userRepo: userRepo}
}

func (h *ProfileHandler) GetLoggedUser(c *fiber.Ctx) error {
	userId, err := auth.MustGetLoggedUserId(c)
	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	dbUser, err := h.userRepo.FindUserByID(c.Context(), *userId)
	if err != nil {
		log.Error("error finding user. ", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	ghClient := gh.NewDefaultGithubClient(dbUser.AccessToken)
	ghUser, _, err := ghClient.Users.Get(c.Context(), "")
	if err != nil {
		log.Error("error getting github user. ", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	return c.JSON(ghUser)
}
