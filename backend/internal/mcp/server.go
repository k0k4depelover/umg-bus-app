package mcp

/*
================================================================================
MCP SERVER - TRACKING DE UBICACIONES
================================================================================

Este servidor implementa un patrón tipo "Tool Dispatcher":

- Recibe una request JSON
- Decide qué herramienta ejecutar (switch)
- Procesa input dinámico
- Devuelve respuesta estructurada tipo MCP

Esto es MUY parecido a cómo funcionan los "tools" en sistemas con LLMs.

================================================================================
*/

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/umg-bus-app/backend/internal/location"
	"github.com/umg-bus-app/backend/internal/repository"
)

/*
MCPRequest:
- Tool: define QUÉ función ejecutar
- Input: JSON crudo (RawMessage) que se parsea dinámicamente

RawMessage evita acoplar todos los posibles inputs en un solo struct gigante
*/
type MCPRequest struct {
	Tool  string          `json:"tool"`
	Input json.RawMessage `json:"input"`
}

/*
Respuesta MCP estructurada:
Esto permite que el cliente interprete distintos tipos de contenido
(no solo texto, podría ser JSON, markdown, etc.)
*/
type MCPResponse struct {
	Content []McpContent `json:"content"`
}

type McpContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

/*
Server contiene dependencias:

- hub: sistema en tiempo real (probablemente Redis + WebSockets)
- campusRepo: acceso a DB
*/
type Server struct {
	hub        *location.Hub
	campusRepo *repository.CampusRepo
}

/*
Constructor (inyección de dependencias)
*/
func NewServer(hub *location.Hub, campusRepo *repository.CampusRepo) *Server {
	return &Server{hub: hub, campusRepo: campusRepo}
}

/*
Handle:
Entry point HTTP → convierte request en ejecución MCP
*/
func (s *Server) Handle(c *fiber.Ctx) error {
	var req MCPRequest

	/*
	   BodyParser:
	   - Lee el body HTTP
	   - Lo convierte automáticamente a struct
	   - Internamente usa json.Unmarshal
	*/
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "request invalido"})
	}

	var result string
	var err error

	/*
	   switch en Go:
	   - No necesita break
	   - Evalúa igualdad directa
	   - Es más limpio que múltiples if
	*/
	switch req.Tool {

	case "get_live_location":
		result, err = s.getLiveLocation(c.Context(), req.Input)

	case "get_campus_list":
		result, err = s.getCampusList(c.Context())

	case "get_route_history":
		result, err = s.getRouteHistory(c.Context(), req.Input)

	default:
		return c.Status(400).JSON(fiber.Map{
			"error": fmt.Errorf("tool desconocido: %s", req.Tool),
		})
	}

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	/*
	   Se construye respuesta MCP

	   Nota:
	   - Siempre devolvemos string (aunque sea JSON)
	   - El cliente decide cómo interpretarlo
	*/
	return c.JSON(MCPResponse{
		Content: []McpContent{
			{Type: "text", Text: result},
		},
	})
}

/*
getLiveLocation:

- Convierte input JSON → struct (Unmarshal)
- Consulta al hub (tiempo real)
- Convierte resultado → JSON string (MarshalIndent)
*/
func (s *Server) getLiveLocation(ctx context.Context, input json.RawMessage) (string, error) {

	var params struct {
		CampusID string `json:"campus_id"`
	}

	/*
	   json.Unmarshal:
	   - Convierte []byte → struct
	   - Usa reflection + tags json
	   - Necesita puntero (&params)
	*/
	if err := json.Unmarshal(input, &params); err != nil {
		return "", err
	}

	loc, err := s.hub.GetLiveLocation(ctx, params.CampusID)
	if err != nil {
		// error controlado
		return "No hay piloto activo en este campus", nil
	}

	/*
	   json.MarshalIndent:
	   - Convierte struct → JSON legible
	   - "" = sin prefijo
	   - "  " = indentación de 2 espacios
	*/
	out, _ := json.MarshalIndent(loc, "", "  ")

	return string(out), nil
}

/*
getCampusList:
Consulta DB → devuelve lista en formato JSON
*/
func (s *Server) getCampusList(ctx context.Context) (string, error) {
	campuses, err := s.campusRepo.GetAll(ctx)
	if err != nil {
		return "", err
	}

	out, _ := json.MarshalIndent(campuses, "", "  ")
	return string(out), nil
}

/*
getRouteHistory:

- Parsea input dinámico
- Ejecuta query manual
- Itera filas (cursor)
- Construye slice
*/
func (s *Server) getRouteHistory(ctx context.Context, input json.RawMessage) (string, error) {

	var params struct {
		CampusID string `json:"campus_id"`
		Limit    int    `json:"limit"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return "", err
	}

	/*
	   Validación defensiva:
	   - Evita queries enormes
	   - Protege performance
	*/
	if params.Limit == 0 || params.Limit > 500 {
		params.Limit = 100
	}

	rows, err := s.campusRepo.DB().Query(ctx, `
        SELECT lat, lng, bearing, speed_kmh, recorded_at
        FROM location_log
        WHERE campus_id = $1
        ORDER BY recorded_at DESC
        LIMIT $2
    `, params.CampusID, params.Limit)
	if err != nil {
		return "", err
	}

	/*
	   defer:
	   - Se ejecuta al final de la función
	   - Muy usado para liberar recursos
	*/
	defer rows.Close()

	type Point struct {
		Lat        float64 `json:"lat"`
		Lng        float64 `json:"lng"`
		Bearing    float64 `json:"bearing"`
		SpeedKmh   float64 `json:"speed_kmh"`
		RecordedAt string  `json:"recorded_at"`
	}

	var points []Point

	/*
	   rows.Next():
	   - Itera fila por fila
	   - Funciona como cursor
	*/
	for rows.Next() {
		var p Point

		/*
		   Scan:
		   - Copia columnas → variables
		   - Orden importa (igual que SELECT)
		*/
		rows.Scan(&p.Lat, &p.Lng, &p.Bearing, &p.SpeedKmh, &p.RecordedAt)

		points = append(points, p)
	}

	out, _ := json.MarshalIndent(points, "", "  ")
	return string(out), nil
}
