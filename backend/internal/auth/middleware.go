package auth

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

func Middleware(jwtSvc *JWTService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer") {
			return c.Status(401).JSON(fiber.Map{"error": "token required"})
		}
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := jwtSvc.Verify(tokenStr)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "token invalido"})
		}
		c.Locals("claims", claims)
		return c.Next()
	}
}

func RequireRole(role string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, ok := c.Locals("claims").(*Claims)
		if !ok || claims.Role != role {
			return c.Status(403).JSON(fiber.Map{"error": "acceso denegado"})
		}
		return c.Next()
	}
}
