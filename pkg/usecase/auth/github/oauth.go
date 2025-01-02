package github

import (
	"context"
	"driftive.cloud/api/pkg/config"
	"driftive.cloud/api/pkg/db"
	"driftive.cloud/api/pkg/model/auth"
	"driftive.cloud/api/pkg/model/auth/github"
	"driftive.cloud/api/pkg/repository"
	"driftive.cloud/api/pkg/repository/queries"
	"driftive.cloud/api/pkg/usecase/utils/gh"
	"driftive.cloud/api/pkg/usecase/utils/jwt"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	codeExchangeRequestErr        = "error exchanging gh code"
	codeExchangeResponseReadErr   = "error reading gh code exchange response"
	codeExchangePrepareRequestErr = "error preparing gh code exchange request"
)

type OAuthHandler struct {
	cfg            config.Config
	db             *db.DB
	userRepository repository.UserRepository
}

func NewOAuthHandler(cfg config.Config, db *db.DB, userRepo repository.UserRepository) OAuthHandler {
	return OAuthHandler{cfg: cfg, db: db, userRepository: userRepo}
}

func (o *OAuthHandler) Authenticate(c *fiber.Ctx) error {
	state := "test_state"
	authUrl := fmt.
		Sprintf("%s/login/oauth/authorize?client_id=%s&redirect_uri=%s&state=%s",
			o.cfg.GithubAppConfig.GithubURL, o.cfg.GithubAppConfig.ClientID, o.cfg.GithubAppConfig.CallbackURL, state)

	return c.Redirect(authUrl, fiber.StatusFound)
}

func (o *OAuthHandler) Callback(c *fiber.Ctx) error {
	ctx := c.Context()
	code := c.Query("code")
	log.Info("gh auth code: ", code)

	epoch := time.Now().Unix()

	tokenResponse, err := o.ExchangeCodeForToken(code)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	ghClient := gh.NewDefaultGithubClient(tokenResponse.AccessToken)
	user, _, err := ghClient.Users.Get(c.Context(), "")
	if err != nil {
		log.Errorf("Failed to get user: %v", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	log.Info("gh user: ", user)

	err = o.db.WithTx(ctx, func(ctx context.Context) error {
		if err != nil {
			return err
		}

		accessTokenExpiresAt := time.Unix(epoch+int64(tokenResponse.ExpiresIn), 0)
		refreshTokenExpiresAt := time.Unix(epoch+int64(tokenResponse.RefreshTokenExpiresIn), 0)

		createUserParams := queries.CreateOrUpdateUserParams{
			Provider:              "GITHUB",
			ProviderID:            fmt.Sprintf("%d", user.GetID()),
			Name:                  user.GetName(),
			Username:              user.GetLogin(),
			Email:                 user.GetEmail(),
			AccessToken:           tokenResponse.AccessToken,
			AccessTokenExpiresAt:  &accessTokenExpiresAt,
			RefreshToken:          tokenResponse.RefreshToken,
			RefreshTokenExpiresAt: &refreshTokenExpiresAt,
		}

		_, err = o.userRepository.CreateOrUpdateUser(ctx, createUserParams)
		if err != nil {
			return err
		}
		args := queries.FindUserByProviderAndProviderIdParams{
			Provider:   "GITHUB",
			ProviderID: fmt.Sprintf("%d", user.GetID()),
		}
		existingUser, err := o.userRepository.FindUserByProviderAndProviderId(ctx, args)
		if err != nil {
			log.Error("error finding user by provider and provider id: ", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		userToken := auth.UserToken{
			ID:       existingUser.ID,
			Provider: "GITHUB",
		}

		jwtToken, err := jwt.GenerateJWTToken(userToken, o.cfg.Auth.JwtSecret)
		if err != nil {
			log.Error("error generating jwt token: ", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		return c.Redirect(
			fmt.Sprintf("%s?token=%s", o.cfg.Auth.LoginRedirectUrl, jwtToken),
			fiber.StatusFound,
		)
	})

	if err != nil {
		log.Error("error authenticating user: ", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return nil
}

func (o *OAuthHandler) ExchangeCodeForToken(oauthCode string) (*github.AccessTokenResponse, error) {
	client := http.Client{}

	ghUrl := os.Getenv("GITHUB_URL")
	ghClientId := os.Getenv("GITHUB_APP_CLIENT_ID")
	ghClientSecret := os.Getenv("GITHUB_APP_CLIENT_SECRET")

	url := fmt.Sprintf("%s/login/oauth/access_token?client_id=%s&client_secret=%s&code=%s",
		ghUrl, ghClientId, ghClientSecret, oauthCode)

	request, err := http.NewRequest(http.MethodPost, url, nil)
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	if err != nil {
		log.Error("error creating request: ", err)
		return nil, errors.New(codeExchangePrepareRequestErr)
	}
	resp, err := client.Do(request)
	if err != nil {
		log.Error("error sending request: ", err)
		return nil, errors.New(codeExchangeRequestErr)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("error reading response body: ", err)
		return nil, errors.New(codeExchangeResponseReadErr)
	}

	log.Info("gh code exchange response: ", string(respBody))

	tokenResponse := github.AccessTokenResponse{}
	err = json.Unmarshal(respBody, &tokenResponse)
	if err != nil {
		log.Error("error unmarshalling response: ", err)
		return nil, err
	}

	return &tokenResponse, nil
}
