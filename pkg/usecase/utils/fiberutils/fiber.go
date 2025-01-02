package fiberutils

import (
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func GetUserId(c *fiber.Ctx) int64 {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	if claims == nil || claims["user_id"] == nil {
		return 0
	}
	userIdInt64 := int64(claims["user_id"].(float64))
	return userIdInt64
}
