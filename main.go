package main

import (
	"context"
	"driftive.cloud/api/pkg/config"
	"driftive.cloud/api/pkg/db"
	"driftive.cloud/api/pkg/middleware/perms"
	"driftive.cloud/api/pkg/model"
	"driftive.cloud/api/pkg/repository"
	"driftive.cloud/api/pkg/usecase/auth"
	"driftive.cloud/api/pkg/usecase/auth/github"
	"driftive.cloud/api/pkg/usecase/drift_stream"
	"driftive.cloud/api/pkg/usecase/orgs"
	"driftive.cloud/api/pkg/usecase/repos"
	github3 "driftive.cloud/api/pkg/usecase/sync/org/github"
	github2 "driftive.cloud/api/pkg/usecase/sync/user_resources/github"
	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/joho/godotenv"
	"strconv"
	"time"
)

func jwtError(c *fiber.Ctx, err error) error {
	if err.Error() == "Missing or malformed JWT" {
		return c.Status(fiber.StatusBadRequest).
			JSON(fiber.Map{"status": "error", "message": "Missing or malformed JWT", "data": nil})
	}
	return c.Status(fiber.StatusUnauthorized).
		JSON(fiber.Map{"status": "error", "message": "Invalid or expired JWT", "data": nil})
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// load config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Panic("error loading configs. ", err)
	}
	db_ := db.NewDB(*cfg)

	err = db_.Pool.Ping(context.Background())
	if err != nil {
		log.Panic("error connecting to database. ", err)
	}
	repo := repository.NewRepository(db_, cfg)

	app := fiber.New()
	app.Use(requestid.New())
	app.Use(logger.New(logger.Config{
		TimeFormat: time.RFC3339,
		Format:     "${time} | ${locals:requestid} | ${status} | ${latency} | ${ip} | ${method} | ${path} | ${error}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
	}))
	app.Use(healthcheck.New())
	app.Use(compress.New())

	api := app.Group("/api")
	v1 := api.Group("/v1")

	// Repos
	userRepo := repo.UserRepository()
	orgRepo := repo.GitOrgRepository()
	repoRepo := repo.GitRepoRepository()
	syncStatusUserRepo := repo.SyncStatusUserRepository()
	driftRepo := repo.DriftAnalysisRepository()
	orgSyncRepo := repo.GitOrgSyncRepository()

	// syncers
	orgSync := github3.NewSyncOrganization(orgRepo, repoRepo, orgSyncRepo)
	ghTokenRefresher := github.NewTokenRefresher(*cfg, userRepo)
	syncer := github2.NewUserResourceSyncer(userRepo, orgRepo, repoRepo, syncStatusUserRepo)

	// handlers
	ghOAuthHandler := github.NewOAuthHandler(*cfg, db_, userRepo, syncStatusUserRepo)
	organizationHandler := orgs.NewGitOrganizationHandler(*cfg, db_, orgRepo)
	repositoryHandler := repos.NewGitRepositoryHandler(orgRepo, repoRepo, userRepo)
	driftStateHandler := drift_stream.NewDriftStateHandler(orgRepo, repoRepo, driftRepo)
	profileHandler := auth.NewProfileHandler(userRepo)

	// Public routes
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})
	v1.Get("/auth/github", func(c *fiber.Ctx) error {
		return ghOAuthHandler.Authenticate(c)
	})
	v1.Get("/auth/github/callback", func(c *fiber.Ctx) error {
		return ghOAuthHandler.Callback(c)
	})
	v1.Post("/drift_analysis", func(c *fiber.Ctx) error { return driftStateHandler.HandleUpdate(c) })

	app.Use(jwtware.New(jwtware.Config{
		SigningKey:   jwtware.SigningKey{Key: []byte(cfg.Auth.JwtSecret)},
		ErrorHandler: jwtError,
	}))
	app.Use(perms.New(orgRepo))

	// Authenticated routes
	v1.Get("/auth/me", func(c *fiber.Ctx) error { return profileHandler.GetLoggedUser(c) })
	v1.Get("/org/:org_id/repos", func(c *fiber.Ctx) error { return repositoryHandler.ListOrganizationRepos(c) })
	v1.Get("/org/:org_id/repo", func(c *fiber.Ctx) error { return repositoryHandler.GetRepoByOrgIdAndName(c) })
	v1.Get("/repo/:repo_id/token", func(c *fiber.Ctx) error { return repositoryHandler.GetRepoTokenById(c) })
	v1.Post("/repo/:repo_id/token", func(c *fiber.Ctx) error { return repositoryHandler.RegenerateToken(c) })
	v1.Get("/repo/:repo_id/runs", func(c *fiber.Ctx) error { return driftStateHandler.ListRunsByRepoId(c) })
	v1.Get("/analysis/run/:run_id", func(c *fiber.Ctx) error { return driftStateHandler.GetRunById(c) })

	ghG := v1.Group("/gh")
	ghG.Get("/orgs", func(c *fiber.Ctx) error { return organizationHandler.ListGitOrganizations(c) })
	ghG.Get("/org", func(c *fiber.Ctx) error { return organizationHandler.GetOrgByNameAndProvider(c, model.GitHubProvider) })

	ghG.Get("/orgs/sync", func(c *fiber.Ctx) error {
		log.Info("syncing org by id")
		orgIdStr := c.Query("org_id")
		orgIdInt64, err := strconv.ParseInt(orgIdStr, 10, 64)
		if err != nil {
			log.Error("error parsing org_id. ", err)
			return c.SendStatus(fiber.StatusBadRequest)
		}
		go orgSync.SyncOrganizationRepositories(c.Context(), orgIdInt64)
		return c.SendStatus(fiber.StatusOK)
	})

	// Start background jobs
	go ghTokenRefresher.RefreshTokens()
	go syncer.StartSyncLoop()
	go orgSync.StartSyncLoop()

	err = app.Listen(":3000")
	if err != nil {
		log.Panic("error starting server. ", err)
	}
}
