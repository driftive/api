package gh

import (
	"github.com/google/go-github/v67/github"
	"net/http"
)

func NewDefaultGithubClient(token string) *github.Client {
	httpClient := http.Client{}
	ghClient := github.NewClient(&httpClient)

	if token != "" {
		return ghClient.WithAuthToken(token)
	}

	return ghClient
}
