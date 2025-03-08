package github

import (
	"context"
	"driftive.cloud/api/pkg/config"
	"driftive.cloud/api/pkg/model/auth/github"
	"driftive.cloud/api/pkg/repository"
	"driftive.cloud/api/pkg/repository/queries"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2/log"
	"resty.dev/v3"
	"time"
)

type TokenRefresher struct {
	cfg            config.Config
	userRepository repository.UserRepository
}

func NewTokenRefresher(cfg config.Config, userRepository repository.UserRepository) *TokenRefresher {
	return &TokenRefresher{
		cfg:            cfg,
		userRepository: userRepository,
	}
}

type RefreshTokenBody struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RefreshToken string `json:"refresh_token"`
	GrantType    string `json:"grant_type"`
}

const (
	refreshTokenRequestErr = "error refreshing token"
)

func (r *TokenRefresher) SendHttpReq(ctx context.Context, body RefreshTokenBody) (*github.AccessTokenResponse, error) {
	client := resty.New()
	defer client.Close()

	resp, err := client.R().
		WithContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		SetResult(&github.AccessTokenResponse{}).
		Post(fmt.Sprintf("%s/login/oauth/access_token", r.cfg.GithubAppConfig.GithubURL))
	if err != nil {
		log.Errorf("error sending request: %v", err)
		return nil, errors.New(refreshTokenRequestErr)
	}

	if resp.IsError() {
		log.Errorf("error response status: %v", resp.Status())
		return nil, errors.New(refreshTokenRequestErr)
	}
	tokenResponse := resp.Result().(*github.AccessTokenResponse)
	return tokenResponse, nil
}

func (r *TokenRefresher) RefreshToken(ctx context.Context, user *queries.User) error {
	now := time.Now()

	tokenResponse, err := r.SendHttpReq(ctx, RefreshTokenBody{
		ClientId:     r.cfg.GithubAppConfig.ClientID,
		ClientSecret: r.cfg.GithubAppConfig.ClientSecret,
		RefreshToken: user.RefreshToken,
		GrantType:    "refresh_token",
	})
	if err != nil {
		log.Errorf("error refreshing token: %v", err)
		return errors.New(refreshTokenRequestErr)
	}

	var accessTokenExpiresAt time.Time
	var refreshTokenExpiresAt time.Time

	if tokenResponse.ExpiresIn > 0 {
		accessTokenExpiresAt = now.Add(time.Duration(tokenResponse.ExpiresIn) * time.Second)
	}

	if tokenResponse.RefreshTokenExpiresIn > 0 {
		refreshTokenExpiresAt = now.Add(time.Duration(tokenResponse.RefreshTokenExpiresIn) * time.Second)
	}

	_, err = r.userRepository.UpdateUserTokens(ctx, queries.UpdateUserTokensParams{
		ID:                    user.ID,
		AccessToken:           tokenResponse.AccessToken,
		AccessTokenExpiresAt:  &accessTokenExpiresAt,
		RefreshToken:          tokenResponse.RefreshToken,
		RefreshTokenExpiresAt: &refreshTokenExpiresAt,
	})

	if err != nil {
		log.Errorf("error updating user tokens: %v", err)
		return errors.New(refreshTokenRequestErr)
	}

	log.Infof("refreshed token: %v", tokenResponse)
	return nil
}

func (r *TokenRefresher) RefreshTokens() {
	log.Info("starting token refresher")
	var fetchedUsers int
	for {
		log.Info("fetching users with expiring tokens...")
		ctx := context.Background()

		dateParam := time.Now().Add(30 * time.Minute)

		params := queries.FindExpiringTokensByProviderParams{
			Provider:    "GITHUB",
			Queryoffset: 0,
			Maxresults:  20,
			Date:        &dateParam,
		}

		users, err := r.userRepository.FindExpiringTokensByProvider(ctx, params)
		if err != nil {
			log.Errorf("error fetching users: %v", err)
			continue
		}

		fetchedUsers = len(users)

		for _, user := range users {
			err := r.RefreshToken(ctx, &user)
			if err != nil {
				log.Errorf("error refreshing token: %v", err)
			}
		}

		if fetchedUsers == 0 {
			log.Info("sleeping for 10 minutes...")
			time.Sleep(10 * time.Minute)
		}
	}
}
