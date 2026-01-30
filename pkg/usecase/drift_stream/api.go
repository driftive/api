package drift_stream

import (
	"context"
	"driftive.cloud/api/pkg/config"
	"driftive.cloud/api/pkg/model/dto"
	"driftive.cloud/api/pkg/repository"
	"driftive.cloud/api/pkg/repository/queries"
	"driftive.cloud/api/pkg/usecase/cleanup"
	"driftive.cloud/api/pkg/usecase/utils/auth"
	"driftive.cloud/api/pkg/usecase/utils/parsing"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"time"
)

type DriftStateHandler struct {
	cfg                     *config.Config
	orgRepository           repository.GitOrgRepository
	repoRepository          repository.GitRepositoryRepository
	driftAnalysisRepository repository.DriftAnalysisRepository
	cleanupService          *cleanup.CleanupService
}

// DriftAnalysisResponse is the response returned after a successful drift analysis upload
type DriftAnalysisResponse struct {
	RunID        string `json:"run_id"`
	DashboardURL string `json:"dashboard_url"`
}

func NewDriftStateHandler(
	cfg *config.Config,
	orgRepository repository.GitOrgRepository,
	repoRepository repository.GitRepositoryRepository,
	driftAnalysisRepo repository.DriftAnalysisRepository,
	cleanupService *cleanup.CleanupService) *DriftStateHandler {
	return &DriftStateHandler{
		cfg:                     cfg,
		orgRepository:           orgRepository,
		repoRepository:          repoRepository,
		driftAnalysisRepository: driftAnalysisRepo,
		cleanupService:          cleanupService,
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

	// Fetch organization to build dashboard URL
	org, err := d.orgRepository.FindGitOrgById(c.Context(), repo.OrganizationID)
	if err != nil {
		log.Errorf("Error finding organization by ID: %v", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	var state DriftDetectionResult
	if err := c.BodyParser(&state); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	log.Debugf("Received drift state update: %v", state)

	// Use sent value or calculate errored count from project results as fallback
	var totalErrored int32
	if state.TotalErrored != nil {
		totalErrored = *state.TotalErrored
	} else {
		for _, project := range state.ProjectResults {
			if !project.Succeeded {
				totalErrored++
			}
		}
	}

	var runUUID uuid.UUID
	err = d.driftAnalysisRepository.WithTx(c.Context(), func(ctx context.Context) error {
		params := queries.CreateDriftAnalysisRunParams{
			Uuid:                   uuid.New(),
			RepositoryID:           repo.ID,
			TotalProjects:          state.TotalProjects,
			TotalProjectsDrifted:   state.TotalDrifted,
			TotalProjectsErrored:   totalErrored,
			TotalProjectsSkipped:   state.TotalSkipped,
			AnalysisDurationMillis: state.Duration.Milliseconds(),
		}

		run, err := d.driftAnalysisRepository.CreateDriftAnalysisRun(ctx, params)
		if err != nil {
			log.Errorf("Error creating drift analysis run: %v", err)
			return err
		}
		runUUID = run.Uuid

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
				SkippedDueToPr:     project.SkippedDueToPR,
			}
			res, err := d.driftAnalysisRepository.CreateDriftAnalysisProject(ctx, projectParams)
			if err != nil {
				log.Errorf("Error creating drift analysis project: %v", err)
				return err
			}
			log.Debugf("Created drift analysis project: [ID: %d, dir: %s]", res.ID, project.Project.Dir)
		}

		log.Info("Created drift analysis run: ", run.Uuid)
		return nil
	})

	if err != nil {
		log.Errorf("Error handling drift state update: %v", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// Trigger cleanup after successful insert (non-blocking, log errors but don't fail the request)
	if d.cleanupService != nil {
		if cleanupErr := d.cleanupService.CleanupRepositoryRuns(c.Context(), repo.ID); cleanupErr != nil {
			log.Warnf("Cleanup failed for repository %d: %v", repo.ID, cleanupErr)
		}
	}

	// Build dashboard URL: /:provider/:org/:repo/run/:run_uuid
	dashboardURL := fmt.Sprintf("%s/%s/%s/%s/run/%s",
		d.cfg.Frontend.FrontendURL,
		org.Provider,
		org.Name,
		repo.Name,
		runUUID.String(),
	)

	response := DriftAnalysisResponse{
		RunID:        runUUID.String(),
		DashboardURL: dashboardURL,
	}

	return c.JSON(response)
}

func (d *DriftStateHandler) ListRunsByRepoId(c *fiber.Ctx) error {
	userId, err := auth.MustGetLoggedUserId(c)
	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	repoIdStr := c.Params("repo_id")
	if repoIdStr == "" {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	repoId := parsing.StringToInt64(repoIdStr)

	// Check if user is a member of the organization
	isMember, err := d.orgRepository.IsUserMemberOfOrganizationByRepoId(c.Context(), repoId, *userId)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	if !isMember {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

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
	log.Infof("Found %d drift analysis runs for repository ID: %d", len(runsDTO), repoId)
	return c.JSON(runsDTO)
}

func (d *DriftStateHandler) GetRunById(c *fiber.Ctx) error {
	userId, err := auth.MustGetLoggedUserId(c)
	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	runIdStr := c.Params("run_id")
	if runIdStr == "" {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	runId, err := uuid.Parse(runIdStr)
	if err != nil {
		log.Errorf("Error parsing run ID: %v", err)
		return c.SendStatus(fiber.StatusBadRequest)
	}

	run, err := d.driftAnalysisRepository.FindDriftAnalysisRunByUUID(c.Context(), runId)
	if err != nil {
		log.Errorf("Error finding drift analysis run by ID: %v", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	projects, err := d.driftAnalysisRepository.FindDriftAnalysisProjectsByRunId(c.Context(), runId)
	if err != nil {
		log.Errorf("Error finding drift analysis projects by run ID: %v", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// Check if user is a member of the organization
	isMember, err := d.orgRepository.IsUserMemberOfOrganizationByRepoId(c.Context(), run.RepositoryID, *userId)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	if !isMember {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	runDTO := parsing.ToDriftAnalysisRunWithProjectsDTO(run, projects)
	return c.JSON(runDTO)
}

func (d *DriftStateHandler) GetRepositoryStats(c *fiber.Ctx) error {
	userId, err := auth.MustGetLoggedUserId(c)
	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	repoIdStr := c.Params("repo_id")
	if repoIdStr == "" {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	repoId := parsing.StringToInt64(repoIdStr)

	// Check if user is a member of the organization
	isMember, err := d.orgRepository.IsUserMemberOfOrganizationByRepoId(c.Context(), repoId, *userId)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	if !isMember {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	stats, err := d.driftAnalysisRepository.GetRepositoryRunStats(c.Context(), repoId)
	if err != nil {
		log.Errorf("Error getting repository run stats: %v", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	result := dto.RepositoryRunStatsDTO{
		TotalRuns:     stats.TotalRuns,
		RunsWithDrift: stats.RunsWithDrift,
	}

	// Handle last_run_at which can be nil
	if stats.LastRunAt != nil {
		if t, ok := stats.LastRunAt.(time.Time); ok {
			result.LastRunAt = &t
		}
	}

	// Get latest run details if there are runs
	if stats.TotalRuns > 0 {
		latestRun, err := d.driftAnalysisRepository.GetLatestRunForRepository(c.Context(), repoId)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			log.Errorf("Error getting latest run for repository: %v", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		if err == nil {
			runDTO := parsing.ToDriftAnalysisRunDTO(latestRun)
			result.LatestRun = &runDTO
		}
	}

	return c.JSON(result)
}

func (d *DriftStateHandler) GetRepositoryTrends(c *fiber.Ctx) error {
	userId, err := auth.MustGetLoggedUserId(c)
	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	repoIdStr := c.Params("repo_id")
	if repoIdStr == "" {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	repoId := parsing.StringToInt64(repoIdStr)

	// Check if user is a member of the organization
	isMember, err := d.orgRepository.IsUserMemberOfOrganizationByRepoId(c.Context(), repoId, *userId)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	if !isMember {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	// Parse days_back query param (default 30, max 90)
	daysBack := int32(c.QueryInt("days_back", 30))
	if daysBack < 1 {
		daysBack = 30
	}
	if daysBack > 90 {
		daysBack = 90
	}

	// Fetch drift rate over time
	driftRate, err := d.driftAnalysisRepository.GetDriftRateOverTime(c.Context(), repoId, daysBack)
	if err != nil {
		log.Errorf("Error getting drift rate: %v", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// Fetch most frequently drifted projects
	frequentlyDrifted, err := d.driftAnalysisRepository.GetMostFrequentlyDriftedProjects(c.Context(), repoId, daysBack, 10)
	if err != nil {
		log.Errorf("Error getting frequently drifted projects: %v", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// Fetch drift-free streak
	var streakDTO dto.DriftFreeStreakDTO
	streak, err := d.driftAnalysisRepository.GetDriftFreeStreak(c.Context(), repoId)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			log.Errorf("Error getting drift-free streak: %v", err)
		}
		streakDTO = dto.DriftFreeStreakDTO{StreakCount: 0}
	} else {
		streakDTO = parsing.ToDriftFreeStreakDTO(streak)
	}

	// Fetch mean time to resolution
	resolutionTimes, err := d.driftAnalysisRepository.GetMeanTimeToResolution(c.Context(), repoId, daysBack)
	if err != nil {
		log.Errorf("Error getting resolution times: %v", err)
		resolutionTimes = nil
	}

	// Calculate summary statistics
	driftRateData := parsing.ToDriftRateDataPoints(driftRate)
	var totalRuns int64
	var runsWithDrift int64
	for _, dp := range driftRateData {
		totalRuns += dp.TotalRuns
		runsWithDrift += dp.RunsWithDrift
	}

	driftRatePercent := float64(0)
	if totalRuns > 0 {
		driftRatePercent = float64(runsWithDrift) / float64(totalRuns) * 100
	}

	summary := dto.TrendsSummaryDTO{
		TotalRuns:        totalRuns,
		RunsWithDrift:    runsWithDrift,
		DriftRatePercent: driftRatePercent,
		StreakCount:      streakDTO.StreakCount,
	}

	response := dto.RepositoryTrendsDTO{
		Summary:                   summary,
		DriftRateOverTime:         driftRateData,
		FrequentlyDriftedProjects: parsing.ToFrequentlyDriftedProjects(frequentlyDrifted),
		DriftFreeStreak:           streakDTO,
		ResolutionTimes:           parsing.ToResolutionTimeDataPoints(resolutionTimes),
		DaysBack:                  int(daysBack),
	}

	return c.JSON(response)
}
