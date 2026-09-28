package respond

import "github.com/gofiber/fiber/v2"

func Error(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{"error": message})
}
