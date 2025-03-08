package auth

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/golang-jwt/jwt/v5"
)

const (
	UserIdKey              = "user_id"
	ErrUserNotFoundMsg     = "E0001_Unauthorized"
	ErrOrgNotAuthorizedMsg = "E0002_Unauthorized"
)

func MustGetLoggedUserId(c *fiber.Ctx) (*int64, error) {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userIdInt64 := int64(claims[UserIdKey].(float64))
	if userIdInt64 == 0 {
		return nil, fiber.NewError(fiber.StatusUnauthorized, ErrUserNotFoundMsg)
	}
	return &userIdInt64, nil
}

func MustHavePermission(c *fiber.Ctx, orgId int64) error {
	userOrgIds := c.Locals("user_org_ids").([]int64)
	for _, id := range userOrgIds {
		if id == orgId {
			log.Debug("user has permission to access org: ", orgId)
			return nil
		}
	}
	return fiber.NewError(fiber.StatusUnauthorized, ErrOrgNotAuthorizedMsg)
}
