package auth_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/umg-bus-app/backend/internal/auth"
	"github.com/umg-bus-app/backend/internal/testutil"
)

func setupAuthApp(env *testutil.TestEnv) *fiber.App {
	app := fiber.New()
	handler := auth.NewHandler(env.PG, env.Redis, env.JWT)
	app.Post("/auth/login", handler.Login)
	app.Post("/auth/refresh", handler.Refresh)
	app.Post("/auth/logout", handler.Logout)
	return app
}

func postJSON(app *fiber.App, path string, body any) (*http.Response, []byte) {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)
	data, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	return resp, data
}

// --- Login tests ---

func TestLogin_Pilot_Success(t *testing.T) {
	env := testutil.Setup(t)
	app := setupAuthApp(env)

	resp, body := postJSON(app, "/auth/login", map[string]string{
		"username": "testpilot",
		"password": "pilotpass123",
		"role":     "pilot",
	})
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, body)
	}

	var result auth.LoginResponse
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if result.AccessToken == "" {
		t.Error("expected non-empty access_token")
	}
	if result.RefreshToken == "" {
		t.Error("expected non-empty refresh_token")
	}

	// Verify the access token contains correct claims
	claims, err := env.JWT.Verify(result.AccessToken)
	if err != nil {
		t.Fatalf("access token should be valid: %v", err)
	}
	if claims.UserID != env.PilotID {
		t.Errorf("expected UserID=%s, got %s", env.PilotID, claims.UserID)
	}
	if claims.Role != "pilot" {
		t.Errorf("expected Role=pilot, got %s", claims.Role)
	}
}

func TestLogin_Student_Success(t *testing.T) {
	env := testutil.Setup(t)
	app := setupAuthApp(env)

	resp, body := postJSON(app, "/auth/login", map[string]string{
		"username": "teststudent",
		"password": "studentpass123",
		"role":     "student",
	})
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, body)
	}

	var result auth.LoginResponse
	json.Unmarshal(body, &result)
	if result.AccessToken == "" || result.RefreshToken == "" {
		t.Error("expected non-empty tokens")
	}

	claims, _ := env.JWT.Verify(result.AccessToken)
	if claims.UserID != env.StudentID {
		t.Errorf("expected UserID=%s, got %s", env.StudentID, claims.UserID)
	}
	if claims.Role != "student" {
		t.Errorf("expected Role=student, got %s", claims.Role)
	}
}

func TestLogin_InvalidRole(t *testing.T) {
	env := testutil.Setup(t)
	app := setupAuthApp(env)

	resp, _ := postJSON(app, "/auth/login", map[string]string{
		"username": "testpilot",
		"password": "pilotpass123",
		"role":     "admin",
	})
	if resp.StatusCode != 400 {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestLogin_MissingFields(t *testing.T) {
	env := testutil.Setup(t)
	app := setupAuthApp(env)

	resp, _ := postJSON(app, "/auth/login", map[string]string{})
	if resp.StatusCode != 400 {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestLogin_MalformedJSON(t *testing.T) {
	env := testutil.Setup(t)
	app := setupAuthApp(env)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != 400 {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	env := testutil.Setup(t)
	app := setupAuthApp(env)

	resp, _ := postJSON(app, "/auth/login", map[string]string{
		"username": "testpilot",
		"password": "wrongpassword",
		"role":     "pilot",
	})
	if resp.StatusCode != 401 {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	env := testutil.Setup(t)
	app := setupAuthApp(env)

	resp, _ := postJSON(app, "/auth/login", map[string]string{
		"username": "nonexistent",
		"password": "whatever",
		"role":     "pilot",
	})
	if resp.StatusCode != 401 {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestLogin_InactiveUser(t *testing.T) {
	env := testutil.Setup(t)
	app := setupAuthApp(env)

	resp, _ := postJSON(app, "/auth/login", map[string]string{
		"username": "inactivepilot",
		"password": "inactive123",
		"role":     "pilot",
	})
	if resp.StatusCode != 401 {
		t.Fatalf("expected 401 for inactive user, got %d", resp.StatusCode)
	}
}

// --- Refresh tests ---

func loginAndGetTokens(t *testing.T, app *fiber.App, username, password, role string) auth.LoginResponse {
	t.Helper()
	resp, body := postJSON(app, "/auth/login", map[string]string{
		"username": username,
		"password": password,
		"role":     role,
	})
	if resp.StatusCode != 200 {
		t.Fatalf("login failed: %d %s", resp.StatusCode, body)
	}
	var result auth.LoginResponse
	json.Unmarshal(body, &result)
	return result
}

func TestRefresh_ValidToken(t *testing.T) {
	env := testutil.Setup(t)
	app := setupAuthApp(env)

	tokens := loginAndGetTokens(t, app, "testpilot", "pilotpass123", "pilot")

	resp, body := postJSON(app, "/auth/refresh", map[string]string{
		"refresh_token": tokens.RefreshToken,
	})
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, body)
	}

	var result map[string]string
	json.Unmarshal(body, &result)
	if result["access_token"] == "" {
		t.Error("expected non-empty access_token")
	}
	// New access token should be different from the original
	if result["access_token"] == tokens.AccessToken {
		t.Error("expected a different access token")
	}
}

func TestRefresh_InvalidToken(t *testing.T) {
	env := testutil.Setup(t)
	app := setupAuthApp(env)

	resp, _ := postJSON(app, "/auth/refresh", map[string]string{
		"refresh_token": "random-invalid-token",
	})
	if resp.StatusCode != 401 {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestRefresh_EmptyBody(t *testing.T) {
	env := testutil.Setup(t)
	app := setupAuthApp(env)

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != 400 {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestRefresh_ExpiredFromRedis(t *testing.T) {
	env := testutil.Setup(t)
	app := setupAuthApp(env)

	tokens := loginAndGetTokens(t, app, "testpilot", "pilotpass123", "pilot")

	// Manually delete the refresh token from Redis to simulate expiration
	tokenHash := fmt.Sprintf("%x", sha256.Sum256([]byte(tokens.RefreshToken)))
	env.Redis.Del(t.Context(), "refresh:"+tokenHash)

	// Small delay to ensure Redis delete propagates
	time.Sleep(50 * time.Millisecond)

	resp, _ := postJSON(app, "/auth/refresh", map[string]string{
		"refresh_token": tokens.RefreshToken,
	})
	if resp.StatusCode != 401 {
		t.Fatalf("expected 401 after Redis deletion, got %d", resp.StatusCode)
	}
}

// --- Logout tests ---

func TestLogout_Success(t *testing.T) {
	env := testutil.Setup(t)
	app := setupAuthApp(env)

	tokens := loginAndGetTokens(t, app, "testpilot", "pilotpass123", "pilot")

	resp, body := postJSON(app, "/auth/logout", map[string]string{
		"refresh_token": tokens.RefreshToken,
	})
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, body)
	}

	var result map[string]any
	json.Unmarshal(body, &result)
	if result["ok"] != true {
		t.Errorf("expected ok=true, got %v", result["ok"])
	}

	// Verify refresh token is no longer usable
	resp2, _ := postJSON(app, "/auth/refresh", map[string]string{
		"refresh_token": tokens.RefreshToken,
	})
	if resp2.StatusCode != 401 {
		t.Fatalf("expected 401 after logout, got %d", resp2.StatusCode)
	}
}

func TestLogout_EmptyBody(t *testing.T) {
	env := testutil.Setup(t)
	app := setupAuthApp(env)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != 400 {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestLogout_AlreadyRevoked(t *testing.T) {
	env := testutil.Setup(t)
	app := setupAuthApp(env)

	tokens := loginAndGetTokens(t, app, "testpilot", "pilotpass123", "pilot")

	// Logout once
	postJSON(app, "/auth/logout", map[string]string{
		"refresh_token": tokens.RefreshToken,
	})

	// Logout again (should be idempotent)
	resp, _ := postJSON(app, "/auth/logout", map[string]string{
		"refresh_token": tokens.RefreshToken,
	})
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200 for idempotent logout, got %d", resp.StatusCode)
	}
}
