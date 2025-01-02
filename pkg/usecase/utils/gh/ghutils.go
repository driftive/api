package gh

import (
	"context"
	"driftive.cloud/api/pkg/repository"
	"encoding/base64"
	"github.com/google/go-github/v67/github"
	"github.com/jferrl/go-githubauth"
	"golang.org/x/oauth2"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
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

func NewAppGithubClient() (*github.Client, error) {
	ghAppPrivateKeyBase64 := os.Getenv("GITHUB_APP_PRIVATE_KEY_BASE64")
	appID, _ := strconv.ParseInt(os.Getenv("GITHUB_APP_ID"), 10, 64)

	decoder := base64.NewDecoder(base64.StdEncoding, strings.NewReader(ghAppPrivateKeyBase64))
	privateKey, _ := io.ReadAll(decoder)

	appTokenSource, err := githubauth.NewApplicationTokenSource(appID, privateKey)
	if err != nil {
		return nil, err
	}

	oauth2HttpClient := oauth2.NewClient(context.Background(), appTokenSource)
	ghClient := github.NewClient(oauth2HttpClient)
	return ghClient, nil
}
