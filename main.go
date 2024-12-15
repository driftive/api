package main

import (
	"context"
	"driftive.cloud/api/pkg/config"
	"driftive.cloud/api/pkg/db"
	"driftive.cloud/api/pkg/repository"
	"driftive.cloud/api/pkg/usecase/auth/github"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
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

	ghOAuthHandler := github.NewOAuthHandler(*cfg, db_, repo.UserRepository())

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	v1.Get("/auth/oauth/github", func(c fiber.Ctx) error {
		return ghOAuthHandler.Authenticate(c)
	})

	v1.Get("/auth/oauth/github/callback", func(c fiber.Ctx) error {
		return ghOAuthHandler.Callback(c)
	})

	app.Listen(":3000")
}
