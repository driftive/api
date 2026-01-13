package github

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"driftive.cloud/api/pkg/config"
	"driftive.cloud/api/pkg/db"
	"driftive.cloud/api/pkg/model/auth"
	"driftive.cloud/api/pkg/model/auth/github"
	"driftive.cloud/api/pkg/repository"
	"driftive.cloud/api/pkg/repository/queries"
	"driftive.cloud/api/pkg/usecase/utils/gh"
	"driftive.cloud/api/pkg/usecase/utils/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	gojwt "github.com/golang-jwt/jwt/v5"
)

const (
	codeExchangeRequestErr        = "error exchanging gh code"
	codeExchangeResponseReadErr   = "error reading gh code exchange response"
	codeExchangePrepareRequestErr = "error preparing gh code exchange request"

	nonceLength     = 16
	stateExpiration = 10 * time.Minute
)

// OAuthStateClaims represents the JWT claims for OAuth state
type OAuthStateClaims struct {
	Nonce       string `json:"nonce"`
	RedirectURL string `json:"redirect_url,omitempty"`
	gojwt.RegisteredClaims
}

type OAuthHandler struct {
	cfg                      config.Config
	db                       *db.DB
	userRepository           repository.UserRepository
	syncStatusUserRepository repository.SyncStatusUserRepository
}

func NewOAuthHandler(cfg config.Config, db *db.DB, userRepo repository.UserRepository, syncRepo repository.SyncStatusUserRepository) OAuthHandler {
	return OAuthHandler{cfg: cfg, db: db, userRepository: userRepo, syncStatusUserRepository: syncRepo}
}

// isAllowedRedirectURL validates that a redirect URL is allowed.
// It checks against the configured allowlist of origins.
// Returns true if the URL is allowed, false otherwise.
func (o *OAuthHandler) isAllowedRedirectURL(redirectURL string) bool {
	if redirectURL == "" {
		return true // Empty is allowed, will use default
	}

	parsed, err := url.Parse(redirectURL)
	if err != nil {
		return false
	}

	// Must be absolute URL with http or https scheme
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}

	// Build origin from the redirect URL (scheme + host)
	redirectOrigin := fmt.Sprintf("%s://%s", parsed.Scheme, parsed.Host)

	// Check against default login redirect URL origin
	defaultParsed, err := url.Parse(o.cfg.Auth.LoginRedirectUrl)
	if err == nil {
		defaultOrigin := fmt.Sprintf("%s://%s", defaultParsed.Scheme, defaultParsed.Host)
		if strings.EqualFold(redirectOrigin, defaultOrigin) {
			return true
		}
	}

	// Check against explicitly allowed origins
	for _, allowedOrigin := range o.cfg.Auth.AllowedRedirectOrigins {
		if strings.EqualFold(redirectOrigin, allowedOrigin) {
			return true
		}
	}

	return false
}

func generateNonce() (string, error) {
	b := make([]byte, nonceLength)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// generateStateJWT creates a signed JWT token containing the OAuth state
func (o *OAuthHandler) generateStateJWT(redirectURL string) (string, error) {
	nonce, err := generateNonce()
	if err != nil {
		return "", err
	}

	claims := OAuthStateClaims{
		Nonce:       nonce,
		RedirectURL: redirectURL,
		RegisteredClaims: gojwt.RegisteredClaims{
			ExpiresAt: gojwt.NewNumericDate(time.Now().Add(stateExpiration)),
			IssuedAt:  gojwt.NewNumericDate(time.Now()),
		},
	}

	token := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(o.cfg.Auth.JwtSecret))
}

// validateStateJWT validates the JWT state token and returns the claims
func (o *OAuthHandler) validateStateJWT(tokenString string) (*OAuthStateClaims, error) {
	token, err := gojwt.ParseWithClaims(tokenString, &OAuthStateClaims{}, func(token *gojwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*gojwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(o.cfg.Auth.JwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*OAuthStateClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid state token")
	}

	return claims, nil
}

func (o *OAuthHandler) Authenticate(c *fiber.Ctx) error {
	// Optional redirect URL from query parameter
	redirectURL := c.Query("redirect_url", "")

	// Validate redirect URL against allowlist to prevent open redirect attacks
	if !o.isAllowedRedirectURL(redirectURL) {
		log.Warnf("rejected unauthorized redirect URL: %s", redirectURL)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid redirect_url",
		})
	}

	state, err := o.generateStateJWT(redirectURL)
	if err != nil {
		log.Error("error generating oauth state: ", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	authUrl := fmt.Sprintf("%s/login/oauth/authorize?client_id=%s&redirect_uri=%s&state=%s",
		o.cfg.GithubAppConfig.GithubURL, o.cfg.GithubAppConfig.ClientID, o.cfg.GithubAppConfig.CallbackURL, state)

	return c.Redirect(authUrl, fiber.StatusFound)
}

func (o *OAuthHandler) Callback(c *fiber.Ctx) error {
	ctx := c.Context()

	// Validate OAuth state JWT
	stateToken := c.Query("state")
	if stateToken == "" {
		log.Warn("OAuth state missing")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "missing oauth state",
		})
	}

	stateClaims, err := o.validateStateJWT(stateToken)
	if err != nil {
		log.Warnf("OAuth state validation failed: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid oauth state",
		})
	}

	code := c.Query("code")

	epoch := time.Now().Unix()

	tokenResponse, err := o.ExchangeCodeForToken(ctx, code)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	ghClient := gh.NewDefaultGithubClient(tokenResponse.AccessToken)
	user, _, err := ghClient.Users.Get(c.Context(), "")
	if err != nil {
		log.Errorf("Failed to get user: %v", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	err = o.db.WithTx(ctx, func(ctx context.Context) error {
		if err != nil {
			return err
		}

		accessTokenExpiresAt := time.Unix(epoch+int64(tokenResponse.ExpiresIn), 0)
		refreshTokenExpiresAt := time.Unix(epoch+int64(tokenResponse.RefreshTokenExpiresIn), 0)

		upsertUserParams := queries.UpsertUserOnLoginParams{
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

		_, err = o.userRepository.UpsertUserOnLogin(ctx, upsertUserParams)
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

		_, err = o.syncStatusUserRepository.CreateOrUpdateSyncStatusUser(ctx, existingUser.ID)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				log.Error("error creating sync status user: ", err)
				return c.SendStatus(fiber.StatusInternalServerError)
			}
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

		// Use redirect URL from state if provided and allowed, otherwise use default
		redirectURL := o.cfg.Auth.LoginRedirectUrl
		if stateClaims.RedirectURL != "" {
			// Re-validate redirect URL as defense-in-depth
			if o.isAllowedRedirectURL(stateClaims.RedirectURL) {
				redirectURL = stateClaims.RedirectURL
			} else {
				log.Warnf("rejected redirect URL from state (tampered?): %s", stateClaims.RedirectURL)
			}
		}

		return c.Redirect(
			fmt.Sprintf("%s?token=%s", redirectURL, jwtToken),
			fiber.StatusFound,
		)
	})

	if err != nil {
		log.Error("error authenticating user: ", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return nil
}

func (o *OAuthHandler) ExchangeCodeForToken(ctx context.Context, oauthCode string) (*github.AccessTokenResponse, error) {
	client := http.Client{}

	ghUrl := os.Getenv("GITHUB_URL")
	ghClientId := os.Getenv("GITHUB_APP_CLIENT_ID")
	ghClientSecret := os.Getenv("GITHUB_APP_CLIENT_SECRET")

	url := fmt.Sprintf("%s/login/oauth/access_token?client_id=%s&client_secret=%s&code=%s",
		ghUrl, ghClientId, ghClientSecret, oauthCode)

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
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

	tokenResponse := github.AccessTokenResponse{}
	err = json.Unmarshal(respBody, &tokenResponse)
	if err != nil {
		log.Error("error unmarshalling response: ", err)
		return nil, err
	}

	return &tokenResponse, nil
}
