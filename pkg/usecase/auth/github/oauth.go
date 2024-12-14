package github

import (
	"driftive.cloud/api/pkg/usecase/utils/gh"
	"fmt"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"os"
)

type OAuthHandler struct {
}

func NewOAuthHandler() OAuthHandler {
	return OAuthHandler{}
}

func (o *OAuthHandler) Authenticate(c fiber.Ctx) {
	clientId := os.Getenv("GITHUB_APP_OAUTH_CLIENT_ID")
	redirectUri := os.Getenv("GITHUB_APP_OAUTH_REDIRECT_URI")

	state := "test_state"

	authUrl := fmt.
		Sprintf("https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&state=%s",
			clientId, redirectUri, state)

	err := c.Redirect().
		Status(fiber.StatusFound).
		To(authUrl)

	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
	}
}

func (o *OAuthHandler) Callback(c fiber.Ctx) {
	code := c.Query("code")
	//clientId := os.Getenv("GITHUB_APP_OAUTH_CLIENT_ID")
	//clientSecret := os.Getenv("GITHUB_APP_OAUTH_CLIENT_SECRET")
	log.Info("gh auth code: ", code)

	tokenResponse, err := gh.ExchangeCodeForToken(code)
	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		return
	}

	log.Info("gh token response: ", tokenResponse)

	ghClient := gh.NewDefaultGithubClient(tokenResponse.AccessToken)
	user, _, err := ghClient.Users.Get(c.Context(), "")
	if err != nil {
		log.Errorf("Failed to get user: %v", err)
		c.SendStatus(fiber.StatusInternalServerError)
		return
	}

	log.Info("gh user: ", user)

	err = c.Redirect().Status(fiber.StatusFound).To("https://localhost:3000")
	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		return
	}
}
