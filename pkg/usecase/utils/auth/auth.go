package auth

import (
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

const (
	UserIdKey          = "user_id"
	ErrUserNotFoundMsg = "E0001_Unauthorized"
)

func GetLoggedUserId(c *fiber.Ctx) (*int64, error) {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userIdInt64 := int64(claims[UserIdKey].(float64))
	if userIdInt64 == 0 {
		return nil, fiber.NewError(fiber.StatusUnauthorized, ErrUserNotFoundMsg)
	}
	return &userIdInt64, nil
}
