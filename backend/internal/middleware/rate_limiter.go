package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/umg-bus-app/backend/internal/auth"
)

func limitReached(c *fiber.Ctx) error {
	return c.Status(429).JSON(fiber.Map{"error": "too many requests, try again later"})
}

// AuthLimiter returns a rate limiter for authentication endpoints.
// 20 requests per minute per IP.
func AuthLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:          20,
		Expiration:   1 * time.Minute,
		LimitReached: limitReached,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
	})
}

// PublicLimiter returns a rate limiter for public endpoints.
// 120 requests per minute per IP.
func PublicLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:          120,
		Expiration:   1 * time.Minute,
		LimitReached: limitReached,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
	})
}

// ProtectedLimiter returns a rate limiter for authenticated endpoints.
// 200 requests per minute per user (falls back to IP if no claims).
func ProtectedLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:          200,
		Expiration:   1 * time.Minute,
		LimitReached: limitReached,
		KeyGenerator: func(c *fiber.Ctx) string {
			if claims, ok := c.Locals("claims").(*auth.Claims); ok && claims != nil {
				return claims.UserID
			}
			return c.IP()
		},
	})
}

// MCPLimiter returns a rate limiter for the MCP tool endpoint.
// 60 requests per minute per IP.
func MCPLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:          60,
		Expiration:   1 * time.Minute,
		LimitReached: limitReached,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
	})
}
