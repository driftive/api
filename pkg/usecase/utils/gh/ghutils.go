package gh

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"driftive.cloud/api/pkg/repository"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/go-github/v81/github"
	"github.com/jferrl/go-githubauth"
	"golang.org/x/oauth2"
)

func NewDefaultGithubClient(token string) *github.Client {
	httpClient := http.Client{}
	ghClient := github.NewClient(&httpClient)

	if token != "" {
		return ghClient.WithAuthToken(token)
	}

	return ghClient
}

func NewUserGithubClient(ctx context.Context, userId int64, usersRepository repository.UserRepository) (*github.Client, error) {
	user, err := usersRepository.FindUserByID(ctx, userId)
	if err != nil {
		return nil, err
	}

	return NewDefaultGithubClient(user.AccessToken), nil
}

func NewAppGithubInstallationClient(ctx context.Context, installationId int64) (*github.Client, error) {
	ghAppPrivateKeyBase64 := os.Getenv("GITHUB_APP_PRIVATE_KEY_BASE64")
	appID, _ := strconv.ParseInt(os.Getenv("GITHUB_APP_ID"), 10, 64)

	decoder := base64.NewDecoder(base64.StdEncoding, strings.NewReader(ghAppPrivateKeyBase64))
	privateKey, _ := io.ReadAll(decoder)

	appTokenSource, err := githubauth.NewApplicationTokenSource(appID, privateKey)
	if err != nil {
		log.Error("error creating app token source: ", err)
		return nil, err
	}

	installationTokenSource := githubauth.NewInstallationTokenSource(installationId, appTokenSource)

	oauth2HttpClient := oauth2.NewClient(ctx, installationTokenSource)
	ghClient := github.NewClient(oauth2HttpClient)
	return ghClient, nil
}

func NewAppGithubClient(ctx context.Context) (*github.Client, error) {
	ghAppPrivateKeyBase64 := os.Getenv("GITHUB_APP_PRIVATE_KEY_BASE64")
	appID, _ := strconv.ParseInt(os.Getenv("GITHUB_APP_ID"), 10, 64)

	decoder := base64.NewDecoder(base64.StdEncoding, strings.NewReader(ghAppPrivateKeyBase64))
	privateKey, _ := io.ReadAll(decoder)

	appTokenSource, err := githubauth.NewApplicationTokenSource(appID, privateKey, githubauth.WithApplicationTokenExpiration(5*time.Minute))
	if err != nil {
		log.Error("error creating app token source: ", err)
		return nil, err
	}

	oauth2HttpClient := oauth2.NewClient(ctx, appTokenSource)
	ghClient := github.NewClient(oauth2HttpClient)
	return ghClient, nil
}

func ParseOrgRole(role string) string {
	switch role {
	case "admin":
		return "ADMIN"
	case "member":
		return "MEMBER"
	default:
		return "MEMBER"
	}
}
