package main

import (
	"log"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/umg-bus-app/backend/graph"
	"github.com/umg-bus-app/backend/internal/auth"
	"github.com/umg-bus-app/backend/internal/config"
	"github.com/umg-bus-app/backend/internal/db"
	"github.com/umg-bus-app/backend/internal/location"
	"github.com/umg-bus-app/backend/internal/mcp"
	mw "github.com/umg-bus-app/backend/internal/middleware"
	"github.com/umg-bus-app/backend/internal/repository"
)

// SetupApp creates and configures the Fiber application with all routes and middleware.
// Extracted from main() to enable integration testing with the same app configuration.
func SetupApp(pg *pgxpool.Pool, rdb *redis.Client, jwtSvc *auth.JWTService) *fiber.App {
	campusRepo := repository.NewCampusRepo(pg)
	pilotRepo := repository.NewPilotRepo(pg)
	studentRepo := repository.NewStudentRepo(pg)
	hub := location.NewHub(rdb, pg)

	app := fiber.New()

	// --- Public routes (120 req/min per IP) ---
	public := app.Group("/", mw.PublicLimiter())
	public.Get("/health-check", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "online!"})
	})
	public.Get("/campus", func(c *fiber.Ctx) error {
		campuses, err := campusRepo.GetAll(c.Context())
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(campuses)
	})
	public.Get("/campus/:id", func(c *fiber.Ctx) error {
		campus, err := campusRepo.GetByID(c.Context(), c.Params("id"))
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "campus no encontrado"})
		}
		return c.JSON(campus)
	})
	public.Get("/campus/:id/pilot", func(c *fiber.Ctx) error {
		pilot, err := pilotRepo.GetByCampus(c.Context(), c.Params("id"))
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "piloto no encontrado"})
		}
		return c.JSON(pilot)
	})

	// --- Auth routes (20 req/min per IP) ---
	authHandler := auth.NewHandler(pg, rdb, jwtSvc)
	authGroup := app.Group("/auth", mw.AuthLimiter())
	authGroup.Post("/login", authHandler.Login)
	authGroup.Post("/refresh", authHandler.Refresh)
	authGroup.Post("/logout", authHandler.Logout)

	// --- GraphQL (auth middleware THEN 200 req/min per user) ---
	gqlSrv := handler.NewDefaultServer(
		graph.NewExecutableSchema(graph.Config{
			Resolvers: &graph.Resolver{
				CampusRepo:  campusRepo,
				PilotRepo:   pilotRepo,
				StudentRepo: studentRepo,
				Hub:         hub,
			},
		}),
	)
	app.Use("/graphql", auth.Middleware(jwtSvc))
	app.Use("/graphql", mw.ProtectedLimiter())
	app.Post("/graphql", adaptor.HTTPHandler(gqlSrv))

	// --- Playground (no rate limit, dev tool) ---
	app.Get("/playground", adaptor.HTTPHandler(playground.Handler("GraphQL", "/graphql")))

	// --- MCP routes (60 req/min per IP) ---
	mcpServer := mcp.NewServer(hub, campusRepo)
	app.Post("/mcp", mw.MCPLimiter(), mcpServer.Handle)

	return app
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error cargando configuracion %v", err)
	}
	if err := db.RunMigrations(cfg.PostgresDSN); err != nil {
		log.Fatalf("Migrations fallaron: %v", err)
	}

	pg, err := db.NewPostgres(cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("Error cargando postgresql %v", err)
	}
	defer pg.Close()

	rdb, err := db.NewRedis(cfg.RedisAddr, cfg.RedisPass)
	if err != nil {
		log.Fatalf("No se pudo cargar redis %v", err)
	}
	defer rdb.Close()

	jwtSvc := auth.NewJWTService(cfg.JWTSecret)
	app := SetupApp(pg, rdb, jwtSvc)

	// WebSocket en servidor net/http separado
	// Fiber usa fasthttp que no soporta WebSocket hijack via adaptor
	hub := location.NewHub(rdb, pg)
	wsMux := http.NewServeMux()
	wsMux.HandleFunc("/ws/pilot", location.HandlePilot(hub, jwtSvc))
	wsMux.HandleFunc("/ws/student", location.HandleStudent(hub, jwtSvc))
	go func() {
		log.Printf("WebSocket server en puerto %s", cfg.WSPort)
		if err := http.ListenAndServe(":"+cfg.WSPort, wsMux); err != nil {
			log.Fatalf("WS server error: %v", err)
		}
	}()

	log.Printf("Servidor en el puerto %s", cfg.AppPort)
	log.Fatal(app.Listen(":" + cfg.AppPort))
}
