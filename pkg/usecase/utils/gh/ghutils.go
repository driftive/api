package gh

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v3/log"
	"github.com/google/go-github/v67/github"
	"io"
	"net/http"
	"os"
)

type AccessTokenResponse struct {
	AccessToken           string `json:"access_token"`
	ExpiresIn             int    `json:"expires_in"`
	RefreshToken          string `json:"refresh_token"`
	RefreshTokenExpiresIn int    `json:"refresh_token_expires_in"`
	Scope                 string `json:"scope"`
	TokenType             string `json:"token_type"`
}

func NewDefaultGithubClient(token string) *github.Client {
	httpClient := http.Client{}
	ghClient := github.NewClient(&httpClient)

	if token != "" {
		return ghClient.WithAuthToken(token)
	}

	return ghClient
}

const (
	codeExchangeRequestErr        = "error exchanging gh code"
	codeExchangeResponseReadErr   = "error reading gh code exchange response"
	codeExchangePrepareRequestErr = "error preparing gh code exchange request"
)

func ExchangeCodeForToken(oauthCode string) (*AccessTokenResponse, error) {
	client := http.Client{}

	ghUrl := os.Getenv("GITHUB_URL")
	ghClientId := os.Getenv("GITHUB_APP_OAUTH_CLIENT_ID")
	ghClientSecret := os.Getenv("GITHUB_APP_OAUTH_CLIENT_SECRET")

	url := fmt.Sprintf("%s/login/oauth/access_token?client_id=%s&client_secret=%s&code=%s",
		ghUrl, ghClientId, ghClientSecret, oauthCode)

	request, err := http.NewRequest(http.MethodPost, url, nil)
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

	tokenResponse := AccessTokenResponse{}
	json.Unmarshal(respBody, &tokenResponse)

	return &tokenResponse, nil
}
