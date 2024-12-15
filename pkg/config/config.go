package config

import (
	"os"
	"strconv"
)

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

type Config struct {
	Database        Database
	GithubAppConfig GitHubAppConfig
	Auth            AuthConfig
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
}

func LoadConfig() (*Config, error) {

	port, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		return nil, err
	}

	connections, err := strconv.Atoi(getEnv("DB_CONNECTIONS", "10"))
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
		ClientID:     os.Getenv("GITHUB_APP_OAUTH_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_APP_OAUTH_CLIENT_SECRET"),
		CallbackURL:  os.Getenv("GITHUB_APP_OAUTH_CALLBACK_URL"),
		GithubURL:    os.Getenv("GITHUB_URL"),
	}

	auth := AuthConfig{
		LoginRedirectUrl: getEnv("LOGIN_REDIRECT_URL", "http://localhost:3001/login/success"),
		JwtSecret:        getEnv("JWT_SECRET", ""),
	}

	config := Config{
		Database:        database,
		GithubAppConfig: ghAppConfig,
		Auth:            auth,
	}
	return &config, nil
}
