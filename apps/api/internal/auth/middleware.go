package auth

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shagun101pareek/forgeflow/internal/respond"
)

const userIDLocal = "userID"

func RequireAuth(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		token, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || strings.TrimSpace(token) == "" {
			return respond.Error(c, fiber.StatusUnauthorized, "missing or invalid token")
		}

		userID, err := ParseToken(secret, strings.TrimSpace(token))
		if err != nil {
			return respond.Error(c, fiber.StatusUnauthorized, "missing or invalid token")
		}

		c.Locals(userIDLocal, userID)
		return c.Next()
	}
}

func UserIDFrom(c *fiber.Ctx) (uuid.UUID, bool) {
	userID, ok := c.Locals(userIDLocal).(uuid.UUID)
	if !ok || userID == uuid.Nil {
		return uuid.Nil, false
	}
	return userID, true
}
