package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/umg-bus-app/backend/internal/auth"
	"github.com/umg-bus-app/backend/internal/testutil"
)

func setupTestApp(t *testing.T) (*fiber.App, *testutil.TestEnv) {
	t.Helper()
	env := testutil.Setup(t)
	app := SetupApp(env.PG, env.Redis, env.JWT)
	return app, env
}

// --- Health check ---

func TestHealthCheck_Returns200(t *testing.T) {
	app, _ := setupTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/health-check", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestHealthCheck_ResponseFormat(t *testing.T) {
	app, _ := setupTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/health-check", nil)
	resp, _ := app.Test(req, -1)
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if result["status"] != "online!" {
		t.Errorf("expected status=online!, got %s", result["status"])
	}
}

// --- Campus routes ---

func TestCampusRoutes_GetAll(t *testing.T) {
	app, _ := setupTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/campus", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var campuses []map[string]any
	if err := json.Unmarshal(body, &campuses); err != nil {
		t.Fatalf("expected JSON array: %v", err)
	}
	if len(campuses) < 1 {
		t.Error("expected at least 1 campus from seed data")
	}
}

func TestCampusRoutes_GetByID_NotFound(t *testing.T) {
	app, _ := setupTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/campus/00000000-0000-0000-0000-000000000000", nil)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != 404 {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestCampusRoutes_GetPilot_NotFound(t *testing.T) {
	app, _ := setupTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/campus/00000000-0000-0000-0000-000000000000/pilot", nil)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != 404 {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

// --- Auth rate limiting ---

func TestAuthLogin_RateLimited(t *testing.T) {
	app, _ := setupTestApp(t)

	var lastStatus int
	for i := 0; i < 25; i++ {
		b, _ := json.Marshal(map[string]string{
			"username": "testpilot",
			"password": "wrong",
			"role":     "pilot",
		})
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := app.Test(req, -1)
		resp.Body.Close()
		lastStatus = resp.StatusCode
	}

	// After 20 requests, the 21st+ should be rate limited
	if lastStatus != 429 {
		t.Fatalf("expected 429 after rate limit exceeded, got %d", lastStatus)
	}
}

// --- GraphQL auth ---

func TestGraphQL_RequiresAuth(t *testing.T) {
	app, _ := setupTestApp(t)

	req := httptest.NewRequest(http.MethodPost, "/graphql", bytes.NewReader([]byte(`{"query":"{ campuses { campusID } }"}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)

	if resp.StatusCode != 401 {
		t.Fatalf("expected 401 without token, got %d", resp.StatusCode)
	}
}

func TestGraphQL_WithValidToken(t *testing.T) {
	app, env := setupTestApp(t)

	token, _ := env.JWT.GenerateAccess(env.PilotID, env.CampusID, "pilot")

	body, _ := json.Marshal(map[string]string{
		"query": `{ campuses { campusID name } }`,
	})
	req := httptest.NewRequest(http.MethodPost, "/graphql", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ := app.Test(req, -1)

	// Should pass auth (200 or possibly some GraphQL error, but NOT 401)
	if resp.StatusCode == 401 {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected to pass auth, got 401: %s", respBody)
	}
}

// --- MCP endpoint ---

func TestMCP_Endpoint(t *testing.T) {
	app, _ := setupTestApp(t)

	body, _ := json.Marshal(map[string]any{
		"tool":  "get_campus_list",
		"input": map[string]string{},
	})
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)

	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, respBody)
	}
}

// --- Verify SetupApp returns usable app ---

func TestSetupApp_ReturnsApp(t *testing.T) {
	app, _ := setupTestApp(t)
	if app == nil {
		t.Fatal("expected non-nil fiber.App")
	}

	// Verify auth login works through full app
	b, _ := json.Marshal(map[string]string{
		"username": "testpilot",
		"password": "pilotpass123",
		"role":     "pilot",
	})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("login through SetupApp failed: %d %s", resp.StatusCode, body)
	}

	var result auth.LoginResponse
	json.Unmarshal(body, &result)
	if result.AccessToken == "" {
		t.Error("expected access token from login")
	}
}
