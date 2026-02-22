package jwt

import (
	"errors"
	"time"

	"driftive.cloud/api/pkg/model/auth"
	"github.com/golang-jwt/jwt/v5"
)

const (
	ErrJwtSecretInvalidMessage = "JWT secret must be set and at least 32 characters long"
)

func GenerateJWTToken(userToken auth.UserToken, jwtSecret string) (string, error) {
	if len(jwtSecret) < 32 {
		return "", errors.New(ErrJwtSecretInvalidMessage)
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
