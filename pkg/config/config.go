package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"driftive.cloud/api/pkg/utils"
)

type Config struct {
	Database        Database
	GithubAppConfig GitHubAppConfig
	Auth            AuthConfig
	Frontend        FrontendConfig
}

type Database struct {
	User        string
	Password    string
	Host        string
	Port        int
	Database    string
	Connections int32
}

type GitHubAppConfig struct {
	ClientID     string
	ClientSecret string
	CallbackURL  string
	// GithubURL is the URL to the Github. Default is https://github.com
	GithubURL string
}

type AuthConfig struct {
	// LoginRedirectUrl is the URL to redirect to after login. Default is http://localhost:3001/login/success. It should be the URL of the frontend
	LoginRedirectUrl string
	JwtSecret        string
	// AllowedRedirectOrigins is a list of allowed origins for OAuth redirects.
	// If empty, only LoginRedirectUrl origin is allowed.
	// Example: ["https://app.driftive.io", "http://localhost:3001"]
	AllowedRedirectOrigins []string
}

type FrontendConfig struct {
	// FrontendURL is the URL of the frontend. Default is http://localhost:3001
	FrontendURL string
}

func LoadConfig() (*Config, error) {
	port, err := strconv.Atoi(utils.GetEnvOrDefault("DB_PORT", "5432"))
	if err != nil {
		return nil, err
	}

	connections, err := strconv.Atoi(utils.GetEnvOrDefault("DB_CONNECTIONS", "10"))
	if err != nil {
		return nil, err
	}

	database := Database{
		User:        os.Getenv("DB_USER"),
		Password:    os.Getenv("DB_PASSWORD"),
		Host:        os.Getenv("DB_HOST"),
		Port:        port,
		Database:    os.Getenv("DB_NAME"),
		Connections: int32(connections),
	}

	ghAppConfig := GitHubAppConfig{
		ClientID:     os.Getenv("GITHUB_APP_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_APP_CLIENT_SECRET"),
		CallbackURL:  os.Getenv("GITHUB_APP_CALLBACK_URL"),
		GithubURL:    os.Getenv("GITHUB_URL"),
	}

	// Parse allowed redirect origins from comma-separated env var
	var allowedRedirectOrigins []string
	if originsEnv := os.Getenv("ALLOWED_REDIRECT_ORIGINS"); originsEnv != "" {
		for _, origin := range strings.Split(originsEnv, ",") {
			origin = strings.TrimSpace(origin)
			if origin != "" {
				allowedRedirectOrigins = append(allowedRedirectOrigins, origin)
			}
		}
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if len(jwtSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be set and at least 32 characters long")
	}

	auth := AuthConfig{
		LoginRedirectUrl:       utils.GetEnvOrDefault("LOGIN_REDIRECT_URL", "http://localhost:3001/login/success"),
		JwtSecret:              jwtSecret,
		AllowedRedirectOrigins: allowedRedirectOrigins,
	}

	frontend := FrontendConfig{
		FrontendURL: utils.GetEnvOrDefault("DRIFTIVE_UI_BASE_URL", "http://localhost:3001"),
	}

	config := Config{
		Database:        database,
		GithubAppConfig: ghAppConfig,
		Auth:            auth,
		Frontend:        frontend,
	}

	return &config, nil
}
