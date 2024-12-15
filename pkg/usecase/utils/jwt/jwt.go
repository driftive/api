package jwt

import (
	"driftive.cloud/api/pkg/model/auth"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

const (
	ErrJwtSecretEmptyMessage = "jwt.secret.empty"
)

func GenerateJWTToken(userToken auth.UserToken, jwtSecret string) (string, error) {
	if jwtSecret == "" {
		return "", errors.New(ErrJwtSecretEmptyMessage)
	}

	claims := jwt.MapClaims{
		"user_id": userToken.ID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return t, nil
}
