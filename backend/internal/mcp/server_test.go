package mcp_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/umg-bus-app/backend/internal/location"
	"github.com/umg-bus-app/backend/internal/mcp"
	"github.com/umg-bus-app/backend/internal/repository"
	"github.com/umg-bus-app/backend/internal/testutil"
)

func setupMCPApp(env *testutil.TestEnv) *fiber.App {
	hub := location.NewHub(env.Redis, env.PG)
	campusRepo := repository.NewCampusRepo(env.PG)
	mcpServer := mcp.NewServer(hub, campusRepo)

	app := fiber.New()
	app.Post("/mcp", mcpServer.Handle)
	return app
}

func postMCP(app *fiber.App, body any) (*http.Response, []byte) {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)
	data, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	return resp, data
}

func TestMCP_InvalidJSON(t *testing.T) {
	env := testutil.Setup(t)
	app := setupMCPApp(env)

	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != 400 {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestMCP_UnknownTool(t *testing.T) {
	env := testutil.Setup(t)
	app := setupMCPApp(env)

	resp, _ := postMCP(app, map[string]any{
		"tool":  "nonexistent_tool",
		"input": map[string]string{},
	})
	if resp.StatusCode != 400 {
		t.Fatalf("expected 400 for unknown tool, got %d", resp.StatusCode)
	}
}

func TestMCP_GetCampusList_Success(t *testing.T) {
	env := testutil.Setup(t)
	app := setupMCPApp(env)

	resp, body := postMCP(app, map[string]any{
		"tool":  "get_campus_list",
		"input": map[string]string{},
	})
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, body)
	}

	var result mcp.MCPResponse
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if len(result.Content) == 0 {
		t.Fatal("expected non-empty content")
	}
	if result.Content[0].Type != "text" {
		t.Errorf("expected content type=text, got %s", result.Content[0].Type)
	}
	// The text should contain campus data (JSON string with campus names)
	if result.Content[0].Text == "" {
		t.Error("expected non-empty content text")
	}
}

func TestMCP_GetLiveLocation_NoPilot(t *testing.T) {
	env := testutil.Setup(t)
	app := setupMCPApp(env)

	resp, body := postMCP(app, map[string]any{
		"tool": "get_live_location",
		"input": map[string]string{
			"campus_id": env.CampusID,
		},
	})
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, body)
	}

	var result mcp.MCPResponse
	json.Unmarshal(body, &result)
	// When no pilot is active, should return a message (not an error)
	if len(result.Content) == 0 {
		t.Fatal("expected content in response")
	}
	if result.Content[0].Text == "" {
		t.Error("expected message about no active pilot")
	}
}

func TestMCP_GetRouteHistory_DefaultLimit(t *testing.T) {
	env := testutil.Setup(t)
	app := setupMCPApp(env)

	// Request with limit=0 should default to 100
	resp, body := postMCP(app, map[string]any{
		"tool": "get_route_history",
		"input": map[string]any{
			"campus_id": env.CampusID,
			"limit":     0,
		},
	})
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, body)
	}

	var result mcp.MCPResponse
	json.Unmarshal(body, &result)
	if len(result.Content) == 0 {
		t.Fatal("expected content in response")
	}
}

func TestMCP_GetRouteHistory_ExcessiveLimit(t *testing.T) {
	env := testutil.Setup(t)
	app := setupMCPApp(env)

	// limit=999 should be capped (the handler caps at 500, defaults 100 for 0)
	resp, body := postMCP(app, map[string]any{
		"tool": "get_route_history",
		"input": map[string]any{
			"campus_id": env.CampusID,
			"limit":     999,
		},
	})
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, body)
	}
}

func TestMCP_ResponseFormat(t *testing.T) {
	env := testutil.Setup(t)
	app := setupMCPApp(env)

	resp, body := postMCP(app, map[string]any{
		"tool":  "get_campus_list",
		"input": map[string]string{},
	})
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var result mcp.MCPResponse
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not valid MCP format: %v", err)
	}

	if len(result.Content) == 0 {
		t.Fatal("expected at least one content item")
	}

	for _, item := range result.Content {
		if item.Type == "" {
			t.Error("content item missing type field")
		}
		if item.Text == "" {
			t.Error("content item missing text field")
		}
	}
}
