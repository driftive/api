package auth

import (
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

const (
	UserIdKey              = "user_id"
	ErrUserNotFoundMsg     = "E0001_Unauthorized"
	ErrOrgNotAuthorizedMsg = "E0002_Unauthorized"
)

func MustGetLoggedUserId(c *fiber.Ctx) (*int64, error) {
	userLocal := c.Locals("user")
	if userLocal == nil {
		return nil, fiber.NewError(fiber.StatusUnauthorized, ErrUserNotFoundMsg)
	}

	user, ok := userLocal.(*jwt.Token)
	if !ok || user == nil {
		return nil, fiber.NewError(fiber.StatusUnauthorized, ErrUserNotFoundMsg)
	}

	claims, ok := user.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fiber.NewError(fiber.StatusUnauthorized, ErrUserNotFoundMsg)
	}

	userIdRaw, exists := claims[UserIdKey]
	if !exists {
		return nil, fiber.NewError(fiber.StatusUnauthorized, ErrUserNotFoundMsg)
	}

	userIdFloat, ok := userIdRaw.(float64)
	if !ok {
		return nil, fiber.NewError(fiber.StatusUnauthorized, ErrUserNotFoundMsg)
	}

	userIdInt64 := int64(userIdFloat)
	if userIdInt64 == 0 {
		return nil, fiber.NewError(fiber.StatusUnauthorized, ErrUserNotFoundMsg)
	}

	return &userIdInt64, nil
}

func MustHavePermission(c *fiber.Ctx, orgId int64) error {
	userOrgIdsLocal := c.Locals("user_org_ids")
	if userOrgIdsLocal == nil {
		return fiber.NewError(fiber.StatusUnauthorized, ErrOrgNotAuthorizedMsg)
	}

	userOrgIds, ok := userOrgIdsLocal.([]int64)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, ErrOrgNotAuthorizedMsg)
	}

	for _, id := range userOrgIds {
		if id == orgId {
			return nil
		}
	}
	return fiber.NewError(fiber.StatusUnauthorized, ErrOrgNotAuthorizedMsg)
}
