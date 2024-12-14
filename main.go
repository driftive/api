package main

import (
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
	app := fiber.New()

	api := app.Group("/api")
	v1 := api.Group("/v1")

	ghOAuthHandler := github.NewOAuthHandler()

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	v1.Get("/auth/oauth/github", func(c fiber.Ctx) error {
		ghOAuthHandler.Authenticate(c)
		return nil
	})

	v1.Get("/auth/oauth/github/callback", func(c fiber.Ctx) error {
		ghOAuthHandler.Callback(c)
		return nil
	})

	app.Listen(":3000")
}
