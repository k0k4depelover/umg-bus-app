# Skills — internal/mcp

## What This Package Does
Model Context Protocol (MCP) tool dispatcher. Receives tool invocations via `POST /mcp` and routes them to the appropriate handler.

## Request/Response Format
```json
// Request
{ "tool": "tool_name", "input": { ... } }

// Response
{ "content": [{ "type": "text", "text": "..." }] }
```

## Available Tools

| Tool | Input | What It Does |
|------|-------|-------------|
| `get_campus_list` | *(none)* | Returns all active campuses via `campusRepo.GetAll()` |
| `get_live_location` | `{ "campus_id": "uuid" }` | Returns current bus position via `hub.GetLiveLocation()` |
| `get_route_history` | `{ "campus_id": "uuid", "limit": N }` | Queries `location_log` table (default 100, max 500) |

## Dependencies
```go
type MCPServer struct {
    hub        *location.Hub
    campusRepo *repository.CampusRepo
    db         *pgxpool.Pool          // used for get_route_history direct query
}
```

## When Adding a New Tool
1. Add a new `case "tool_name":` in the `Handle()` method's switch statement
2. Define an input struct if needed, unmarshal from `req.Input`
3. Execute logic and return `MCPResponse` with text content
4. No registration needed — the switch statement IS the tool registry

## Testing
`server_test.go` has 7 integration tests using testcontainers. Tests cover unknown tools, all three existing tools, and response format.
