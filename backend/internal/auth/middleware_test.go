package auth

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func setupMiddlewareApp(jwtSvc *JWTService) *fiber.App {
	app := fiber.New()
	app.Use(Middleware(jwtSvc))
	app.Get("/protected", func(c *fiber.Ctx) error {
		claims := c.Locals("claims").(*Claims)
		return c.JSON(fiber.Map{
			"user_id": claims.UserID,
			"role":    claims.Role,
		})
	})
	return app
}

func TestMiddleware_ValidToken(t *testing.T) {
	svc := newTestJWT()
	app := setupMiddlewareApp(svc)

	token, _ := svc.GenerateAccess("user-1", "campus-1", "pilot")
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, body)
	}
}

func TestMiddleware_MissingHeader(t *testing.T) {
	svc := newTestJWT()
	app := setupMiddlewareApp(svc)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 401 {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestMiddleware_NoBearerPrefix(t *testing.T) {
	svc := newTestJWT()
	app := setupMiddlewareApp(svc)

	token, _ := svc.GenerateAccess("user-1", "campus-1", "pilot")
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Token "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 401 {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestMiddleware_InvalidToken(t *testing.T) {
	svc := newTestJWT()
	app := setupMiddlewareApp(svc)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-garbage-token")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 401 {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestMiddleware_ExpiredToken(t *testing.T) {
	svc := newTestJWT()
	app := setupMiddlewareApp(svc)

	claims := Claims{
		UserID:   "user-1",
		CampusID: "campus-1",
		Role:     "pilot",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte(testSecret))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 401 {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func setupRoleApp(jwtSvc *JWTService, requiredRole string) *fiber.App {
	app := fiber.New()
	app.Use(Middleware(jwtSvc))
	app.Use(RequireRole(requiredRole))
	app.Get("/role-protected", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})
	return app
}

func TestRequireRole_Match(t *testing.T) {
	svc := newTestJWT()
	app := setupRoleApp(svc, "pilot")

	token, _ := svc.GenerateAccess("user-1", "campus-1", "pilot")
	req := httptest.NewRequest(http.MethodGet, "/role-protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestRequireRole_Mismatch(t *testing.T) {
	svc := newTestJWT()
	app := setupRoleApp(svc, "pilot")

	token, _ := svc.GenerateAccess("user-1", "campus-1", "student")
	req := httptest.NewRequest(http.MethodGet, "/role-protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 403 {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
}

func TestRequireRole_NoClaims(t *testing.T) {
	app := fiber.New()
	// Skip auth middleware — no claims set
	app.Use(RequireRole("pilot"))
	app.Get("/role-protected", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/role-protected", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 403 {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
}
