package main

import (
	"context"
	"driftive.cloud/api/pkg/config"
	"driftive.cloud/api/pkg/db"
	"driftive.cloud/api/pkg/repository"
	"driftive.cloud/api/pkg/usecase/auth/github"
	"driftive.cloud/api/pkg/usecase/utils/gh"
	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

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

	api := app.Group("/api")
	v1 := api.Group("/v1")

	userRepo := repo.UserRepository()

	ghOAuthHandler := github.NewOAuthHandler(*cfg, db_, userRepo)
	ghTokenRefresher := github.NewTokenRefresher(*cfg, userRepo)

	// Public routes
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})
	v1.Get("/auth/oauth/github", func(c *fiber.Ctx) error {
		return ghOAuthHandler.Authenticate(c)
	})
	v1.Get("/auth/oauth/github/callback", func(c *fiber.Ctx) error {
		return ghOAuthHandler.Callback(c)
	})

	app.Use(jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(cfg.Auth.JwtSecret)},
	}))

	// Authenticated routes
	v1.Get("/auth/me", func(c *fiber.Ctx) error {
		user := c.Locals("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		log.Info(claims)

		userIdInt64 := int64(claims["user_id"].(float64))
		dbUser, err := repo.UserRepository().FindUserByID(c.Context(), userIdInt64)
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
	})

	ghG := v1.Group("/gh")
	ghG.Get("/orgs", func(c *fiber.Ctx) error {
		return c.SendString("orgs")
	})

	// Start background jobs
	go ghTokenRefresher.RefreshTokens()

	app.Listen(":3000")
}
