package repos

import (
	"driftive.cloud/api/pkg/repository"
	"driftive.cloud/api/pkg/repository/queries"
	"driftive.cloud/api/pkg/usecase/utils/auth"
	"driftive.cloud/api/pkg/usecase/utils/parsing"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type GitRepositoryHandler struct {
	userRepository repository.UserRepository
	repoRepository repository.GitRepositoryRepository
	orgRepository  repository.GitOrgRepository
}

func NewGitRepositoryHandler(
	orgRepository repository.GitOrgRepository,
	repoRepository repository.GitRepositoryRepository,
	userRepository repository.UserRepository,
) *GitRepositoryHandler {
	return &GitRepositoryHandler{
		userRepository: userRepository,
		repoRepository: repoRepository,
		orgRepository:  orgRepository,
	}
}

func (h *GitRepositoryHandler) ListOrganizationRepos(c *fiber.Ctx) error {
	userId, err := auth.GetLoggedUserId(c)
	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	orgIdStr := c.Params("org_id")
	if orgIdStr == "" {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	orgId := parsing.StringToInt64(orgIdStr)

	// Check if user is a member of the organization
	isMember, err := h.orgRepository.IsUserMemberOfOrg(c.Context(), orgId, *userId)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	if !isMember {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	repos, err := h.repoRepository.FindGitReposByOrgId(c.Context(), orgId)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	repoDTOs := parsing.ToGitRepositoryDTOs(repos)
	return c.JSON(repoDTOs)
}

func (h *GitRepositoryHandler) GetRepoByOrgIdAndName(c *fiber.Ctx) error {
	userId, err := auth.GetLoggedUserId(c)
	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	orgIdStr := c.Params("org_id")
	if orgIdStr == "" {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	orgId := parsing.StringToInt64(orgIdStr)

	// Check if user is a member of the organization
	isMember, err := h.orgRepository.IsUserMemberOfOrg(c.Context(), orgId, *userId)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	if !isMember {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	repoName := c.Query("repo_name")
	if repoName == "" {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	repo, err := h.repoRepository.FindGitRepositoryByOrgIdAndName(c.Context(), orgId, repoName)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	repoDTO := parsing.ToGitRepositoryDTO(repo)
	return c.JSON(repoDTO)
}

func (h *GitRepositoryHandler) GetRepoTokenById(c *fiber.Ctx) error {
	userId, err := auth.GetLoggedUserId(c)
	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	repoIdStr := c.Params("repo_id")
	if repoIdStr == "" {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	repoId := parsing.StringToInt64(repoIdStr)

	repo, err := h.repoRepository.FindGitRepositoryById(c.Context(), repoId)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// Check if user is a member of the organization
	isMember, err := h.orgRepository.IsUserMemberOfOrg(c.Context(), repo.OrganizationID, *userId)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	if !isMember {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	tokenResponse := struct {
		Token *string `json:"token"`
	}{
		Token: repo.AnalysisToken,
	}

	return c.JSON(tokenResponse)
}

func (h *GitRepositoryHandler) RegenerateToken(c *fiber.Ctx) error {
	userId, err := auth.GetLoggedUserId(c)
	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	repoIdStr := c.Params("repo_id")
	if repoIdStr == "" {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	repoId := parsing.StringToInt64(repoIdStr)

	// Check if user is a member of the organization
	isMember, err := h.orgRepository.IsUserMemberOfOrganizationByRepoId(c.Context(), repoId, *userId)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	if !isMember {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	repo, err := h.repoRepository.FindGitRepositoryById(c.Context(), repoId)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	token := uuid.New().String()
	params := queries.UpdateRepositoryTokenParams{
		Token: &token,
		ID:    repo.ID,
	}
	newToken, err := h.repoRepository.UpdateRepositoryToken(c.Context(), params)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	tokenResponse := struct {
		Token *string `json:"token"`
	}{
		Token: newToken,
	}
	return c.JSON(tokenResponse)
}
