package github

import (
	"driftive.cloud/api/pkg/repository"
	"github.com/gofiber/fiber/v2"
)

type GitRepositoryHandler struct {
	userRepository *repository.UserRepository
}

func NewGithubRepositoryHandler(userRepository *repository.UserRepository) *GitRepositoryHandler {
	return &GitRepositoryHandler{
		userRepository: userRepository,
	}
}

func (h *GitRepositoryHandler) ListGithubRepositories(c *fiber.Ctx) error {
	//userId := fiberutils.GetUserId(c)
	//userGhClient, err := gh.NewUserGithubClient(c.Context(), userId, *h.userRepository)
	//if err != nil {
	//	return c.SendStatus(fiber.StatusInternalServerError)
	//}
	//
	//opts := github.OrganizationsListOptions{
	//	ListOptions: github.ListOptions{
	//		Page:    0,
	//		PerPage: 100,
	//	},
	//}
	//orgs, _, err := userGhClient.Organizations.List(c.Context(), "", opts)
	//if err != nil {
	//	return c.SendStatus(fiber.StatusInternalServerError)
	//}
	//return c.JSON(orgs)

	return nil

}
