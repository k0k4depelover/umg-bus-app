package middleware

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/umg-bus-app/backend/internal/auth"
)

func dummyHandler(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"ok": true})
}

func fireRequests(t *testing.T, app *fiber.App, method, path string, n int) (okCount, blockedCount int) {
	t.Helper()
	for i := 0; i < n; i++ {
		req := httptest.NewRequest(method, path, nil)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("request %d failed: %v", i, err)
		}
		resp.Body.Close()
		if resp.StatusCode == 200 {
			okCount++
		} else if resp.StatusCode == 429 {
			blockedCount++
		} else {
			t.Fatalf("unexpected status %d on request %d", resp.StatusCode, i)
		}
	}
	return
}

func TestAuthLimiter_AllowsUnderLimit(t *testing.T) {
	app := fiber.New()
	app.Post("/auth/login", AuthLimiter(), dummyHandler)

	ok, blocked := fireRequests(t, app, http.MethodPost, "/auth/login", 20)
	if ok != 20 {
		t.Errorf("expected 20 OK responses, got %d (blocked: %d)", ok, blocked)
	}
}

func TestAuthLimiter_BlocksOverLimit(t *testing.T) {
	app := fiber.New()
	app.Post("/auth/login", AuthLimiter(), dummyHandler)

	ok, blocked := fireRequests(t, app, http.MethodPost, "/auth/login", 25)
	if ok != 20 {
		t.Errorf("expected 20 OK, got %d", ok)
	}
	if blocked != 5 {
		t.Errorf("expected 5 blocked, got %d", blocked)
	}
}

func TestPublicLimiter_BlocksOverLimit(t *testing.T) {
	app := fiber.New()
	app.Get("/campus", PublicLimiter(), dummyHandler)

	ok, blocked := fireRequests(t, app, http.MethodGet, "/campus", 125)
	if ok != 120 {
		t.Errorf("expected 120 OK, got %d", ok)
	}
	if blocked != 5 {
		t.Errorf("expected 5 blocked, got %d", blocked)
	}
}

func TestProtectedLimiter_KeyByUser(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret")

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		tokenStr := c.Get("X-User-Token")
		if tokenStr != "" {
			claims, err := jwtSvc.Verify(tokenStr)
			if err == nil {
				c.Locals("claims", claims)
			}
		}
		return c.Next()
	})
	app.Post("/graphql", ProtectedLimiter(), dummyHandler)

	tokenA, _ := jwtSvc.GenerateAccess("user-A", "campus-1", "pilot")
	tokenB, _ := jwtSvc.GenerateAccess("user-B", "campus-1", "student")

	// User A fires 200 requests (should all pass)
	for i := 0; i < 200; i++ {
		req := httptest.NewRequest(http.MethodPost, "/graphql", nil)
		req.Header.Set("X-User-Token", tokenA)
		resp, _ := app.Test(req, -1)
		resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Fatalf("user-A request %d got %d, expected 200", i, resp.StatusCode)
		}
	}

	// User B should still have their own quota
	req := httptest.NewRequest(http.MethodPost, "/graphql", nil)
	req.Header.Set("X-User-Token", tokenB)
	resp, _ := app.Test(req, -1)
	resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("user-B first request got %d, expected 200 (separate quota)", resp.StatusCode)
	}
}

func TestProtectedLimiter_FallbackToIP(t *testing.T) {
	app := fiber.New()
	// No claims middleware — should fall back to IP
	app.Post("/graphql", ProtectedLimiter(), dummyHandler)

	ok, blocked := fireRequests(t, app, http.MethodPost, "/graphql", 205)
	if ok != 200 {
		t.Errorf("expected 200 OK (IP-based), got %d", ok)
	}
	if blocked != 5 {
		t.Errorf("expected 5 blocked, got %d", blocked)
	}
}

func TestMCPLimiter_BlocksOverLimit(t *testing.T) {
	app := fiber.New()
	app.Post("/mcp", MCPLimiter(), dummyHandler)

	ok, blocked := fireRequests(t, app, http.MethodPost, "/mcp", 65)
	if ok != 60 {
		t.Errorf("expected 60 OK, got %d", ok)
	}
	if blocked != 5 {
		t.Errorf("expected 5 blocked, got %d", blocked)
	}
}

func TestRateLimitResponse_Format(t *testing.T) {
	app := fiber.New()
	app.Post("/test", AuthLimiter(), dummyHandler)

	// Exhaust the limit
	for i := 0; i < 20; i++ {
		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		resp, _ := app.Test(req, -1)
		resp.Body.Close()
	}

	// The 21st should be rate limited
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 429 {
		t.Fatalf("expected 429, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	expected := "too many requests, try again later"
	if result["error"] != expected {
		t.Errorf("expected error=%q, got %q", expected, result["error"])
	}
}
