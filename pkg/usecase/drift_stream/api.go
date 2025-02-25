package drift_stream

import (
	"context"
	"driftive.cloud/api/pkg/repository"
	"driftive.cloud/api/pkg/repository/queries"
	"driftive.cloud/api/pkg/usecase/utils/parsing"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
)

type DriftStateHandler struct {
	repoRepository          repository.GitRepositoryRepository
	driftAnalysisRepository repository.DriftAnalysisRepository
}

func NewDriftStateHandler(repoRepository repository.GitRepositoryRepository, driftAnalysisRepo repository.DriftAnalysisRepository) *DriftStateHandler {
	return &DriftStateHandler{
		repoRepository:          repoRepository,
		driftAnalysisRepository: driftAnalysisRepo,
	}
}

func projectTypeToDBString(projectType ProjectType) (string, error) {
	switch projectType {
	case Terraform:
		return "TERRAFORM", nil
	case Tofu:
		return "TOFU", nil
	case Terragrunt:
		return "TERRAGRUNT", nil
	default:
		return "", errors.New("invalid project type")
	}
}

func (d *DriftStateHandler) HandleUpdate(c *fiber.Ctx) error {
	log.Info("Handling drift state update")
	headers := c.GetReqHeaders()
	tokenArr := headers["X-Token"]

	// token is a string[] so we need to check if it's empty
	if len(tokenArr) == 0 {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	token := tokenArr[0]

	repo, err := d.repoRepository.FindGitRepositoryByToken(c.Context(), token)
	if err != nil {
		log.Errorf("Error finding repository by token: %v", err)
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	log.Infof("Handling drift state update for repository: %s", repo.Name)

	var state DriftDetectionResult
	if err := c.BodyParser(&state); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	log.Debugf("Received drift state update: %v", state)

	d.driftAnalysisRepository.WithTx(c.Context(), func(ctx context.Context) error {
		params := queries.CreateDriftAnalysisRunParams{
			Uuid:                   uuid.New(),
			RepositoryID:           repo.ID,
			TotalProjects:          state.TotalProjects,
			TotalProjectsDrifted:   state.TotalDrifted,
			AnalysisDurationMillis: state.Duration.Milliseconds(),
		}

		run, err := d.driftAnalysisRepository.CreateDriftAnalysisRun(c.Context(), params)
		if err != nil {
			log.Errorf("Error creating drift analysis run: %v", err)
			return err
		}

		for _, project := range state.ProjectResults {

			projectType, err := projectTypeToDBString(project.Project.Type)
			if err != nil {
				log.Errorf("Error converting project type to db string: %v", err)
				return c.SendStatus(fiber.StatusBadRequest)
			}

			projectParams := queries.CreateDriftAnalysisProjectParams{
				DriftAnalysisRunID: run.Uuid,
				Dir:                project.Project.Dir,
				Type:               projectType,
				Drifted:            project.Drifted,
				Succeeded:          project.Succeeded,
				InitOutput:         &project.InitOutput,
				PlanOutput:         &project.PlanOutput,
			}
			res, err := d.driftAnalysisRepository.CreateDriftAnalysisProject(c.Context(), projectParams)
			if err != nil {
				log.Errorf("Error creating drift analysis project: %v", err)
				return err
			}
			log.Debugf("Created drift analysis project: [ID: %d, dir: %s]", res.ID, project.Project.Dir)
		}

		log.Info("Created drift analysis run: ", run.Uuid)
		return nil
	})

	return c.SendStatus(fiber.StatusOK)
}

func (d *DriftStateHandler) ListRunsByRepoId(c *fiber.Ctx) error {
	repoIdStr := c.Params("repo_id")
	if repoIdStr == "" {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	repoId := parsing.StringToInt64(repoIdStr)

	page := c.QueryInt("page")
	if page < 0 {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	runs, err := d.driftAnalysisRepository.FindDriftAnalysisRunsByRepositoryID(c.Context(), repoId, page)
	if err != nil {
		log.Errorf("Error finding drift analysis runs by repository ID: %v", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	runsDTO := parsing.ToDriftAnalysisRunDTOs(runs)
	return c.JSON(runsDTO)
}
