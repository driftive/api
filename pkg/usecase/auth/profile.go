package auth

import (
	"driftive.cloud/api/pkg/repository"
	"driftive.cloud/api/pkg/usecase/utils/gh"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/golang-jwt/jwt/v5"
)

type ProfileHandler struct {
	userRepo repository.UserRepository
}

func NewProfileHandler(userRepo repository.UserRepository) ProfileHandler {
	return ProfileHandler{userRepo: userRepo}
}

func (h *ProfileHandler) GetLoggedUser(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	log.Info(claims)

	userIdInt64 := int64(claims["user_id"].(float64))
	dbUser, err := h.userRepo.FindUserByID(c.Context(), userIdInt64)
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
