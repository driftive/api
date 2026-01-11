package github

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"driftive.cloud/api/pkg/config"
	"driftive.cloud/api/pkg/model/auth/github"
	"driftive.cloud/api/pkg/repository"
	"driftive.cloud/api/pkg/repository/queries"
	"github.com/gofiber/fiber/v2/log"
	"github.com/jackc/pgx/v5"
	"resty.dev/v3"
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
	refreshTokenRequestErr  = "error refreshing token"
	maxRefreshAttempts      = 5
	requestThrottleDelay    = 100 * time.Millisecond
	defaultRateLimitBackoff = 60 * time.Second
)

type RateLimitError struct {
	RetryAfter time.Duration
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limited, retry after %v", e.RetryAfter)
}

func (r *TokenRefresher) SendHttpReq(ctx context.Context, body RefreshTokenBody) (*github.AccessTokenResponse, error) {
	client := resty.New()
	defer client.Close()

	resp, err := client.R().
		WithContext(ctx).
		SetContentType("application/json").
		SetHeader("Accept", "application/json").
		SetBody(body).
		SetExpectResponseContentType("application/json").
		SetResult(github.AccessTokenResponse{}).
		Post(fmt.Sprintf("%s/login/oauth/access_token", r.cfg.GithubAppConfig.GithubURL))
	if err != nil {
		log.Errorf("error sending request: %v", err)
		return nil, errors.New(refreshTokenRequestErr)
	}

	// Handle rate limiting (429 Too Many Requests)
	if resp.StatusCode() == 429 {
		retryAfter := parseRetryAfter(resp.Header().Get("Retry-After"))
		log.Warnf("GitHub rate limit hit, retry after %v", retryAfter)
		return nil, &RateLimitError{RetryAfter: retryAfter}
	}

	if resp.IsError() {
		log.Errorf("error response status: %v", resp.Status())
		return nil, errors.New(refreshTokenRequestErr)
	}
	tokenResponse := resp.Result().(*github.AccessTokenResponse)
	return tokenResponse, nil
}

// parseRetryAfter parses GitHub's Retry-After header.
// GitHub specifies this value in seconds (integer format).
// See: https://docs.github.com/en/rest/using-the-rest-api/rate-limits-for-the-rest-api
func parseRetryAfter(header string) time.Duration {
	if header == "" {
		return defaultRateLimitBackoff
	}

	seconds, err := strconv.Atoi(header)
	if err != nil {
		log.Warnf("unexpected Retry-After header format: %s", header)
		return defaultRateLimitBackoff
	}

	return time.Duration(seconds) * time.Second
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
		// Propagate rate limit errors without counting as failure (not user's fault)
		var rateLimitErr *RateLimitError
		if errors.As(err, &rateLimitErr) {
			return rateLimitErr
		}
		log.Errorf("error refreshing token: %v", err)
		r.handleRefreshFailure(ctx, user)
		return errors.New(refreshTokenRequestErr)
	}
	if tokenResponse == nil || tokenResponse.AccessToken == "" || tokenResponse.RefreshToken == "" {
		log.Errorf("invalid token response for user %d", user.ID)
		r.handleRefreshFailure(ctx, user)
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

	log.Infof("token refreshed for user: %d", user.ID)
	return nil
}

func (r *TokenRefresher) handleRefreshFailure(ctx context.Context, user *queries.User) {
	updatedUser, err := r.userRepository.IncrementTokenRefreshAttempts(ctx, user.ID)
	if err != nil {
		log.Errorf("error incrementing refresh attempts for user %d: %v", user.ID, err)
		return
	}

	if updatedUser.TokenRefreshAttempts >= maxRefreshAttempts {
		log.Warnf("disabling token refresh for user %d after %d failed attempts", user.ID, updatedUser.TokenRefreshAttempts)
		_, err = r.userRepository.DisableTokenRefresh(ctx, user.ID)
		if err != nil {
			log.Errorf("error disabling token refresh for user %d: %v", user.ID, err)
		}
	} else {
		log.Warnf("token refresh failed for user %d (attempt %d/%d)", user.ID, updatedUser.TokenRefreshAttempts, maxRefreshAttempts)
	}
}

func (r *TokenRefresher) RefreshTokens(ctx context.Context) {
	log.Info("starting token refresher")
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	// Run immediately on start, then on ticker
	r.processExpiringTokens(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Info("token refresher shutting down...")
			return
		case <-ticker.C:
			r.processExpiringTokens(ctx)
		}
	}
}

func (r *TokenRefresher) processExpiringTokens(ctx context.Context) {
	log.Info("checking for expiring tokens...")

	dateParam := time.Now().Add(30 * time.Minute)
	var totalProcessed int

	// Process tokens one at a time with row locking for multi-instance safety
	for {
		// Check for cancellation between iterations
		select {
		case <-ctx.Done():
			return
		default:
		}

		processed, err := r.processOneToken(ctx, dateParam)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				// No more tokens to process
				break
			}

			// Handle rate limiting - wait and continue
			var rateLimitErr *RateLimitError
			if errors.As(err, &rateLimitErr) {
				log.Warnf("rate limited, waiting %v before continuing...", rateLimitErr.RetryAfter)
				select {
				case <-ctx.Done():
					return
				case <-time.After(rateLimitErr.RetryAfter):
					continue
				}
			}

			log.Errorf("error processing token: %v", err)
			break
		}

		if processed {
			totalProcessed++

			// Throttle requests to avoid hitting rate limits
			select {
			case <-ctx.Done():
				return
			case <-time.After(requestThrottleDelay):
			}
		}
	}

	if totalProcessed == 0 {
		log.Info("no expiring tokens found")
	} else {
		log.Infof("processed %d expiring tokens", totalProcessed)
	}
}

func (r *TokenRefresher) processOneToken(ctx context.Context, expiryThreshold time.Time) (bool, error) {
	var user queries.User
	var refreshErr error

	err := r.userRepository.WithTx(ctx, func(txCtx context.Context) error {
		var err error
		user, err = r.userRepository.FindAndLockExpiringToken(txCtx, queries.FindAndLockExpiringTokenParams{
			Provider: "GITHUB",
			Date:     &expiryThreshold,
		})
		if err != nil {
			return err
		}

		refreshErr = r.RefreshToken(txCtx, &user)
		return nil // Always commit to release the lock; refresh failure is handled separately
	})

	if err != nil {
		return false, err
	}

	// Propagate rate limit errors to caller for proper handling
	var rateLimitErr *RateLimitError
	if errors.As(refreshErr, &rateLimitErr) {
		return false, rateLimitErr
	}

	if refreshErr != nil {
		log.Errorf("error refreshing token for user %d: %v", user.ID, refreshErr)
	}

	return true, nil
}
