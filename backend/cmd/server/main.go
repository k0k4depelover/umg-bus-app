// Paso 4. Podriamos ya definir el main, sin embargo lo vamos
// Actualizando, asi que el main es un comodin, pero aca añadimos,
// QUE HACEMOS // PASOS PREVIOS NECESARIOS
// 1. Agregar datos del archivo de configuraciones. (.env) // Paso 0
// 2. Llamadas a conectores de la base de datos. // Paso 2 y 3
// 3. Logs de status, endpoints de status // Paso 2 y 3
// 4. Añadir los archivos de migracion // Paso 4
// 5. Añadir el hub // Paso

package main

import (
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/umg-bus-app/backend/internal/config"
	"github.com/umg-bus-app/backend/internal/db"
	"github.com/umg-bus-app/backend/internal/location"
	"github.com/umg-bus-app/backend/internal/repository"
)

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
	// Se ejecuta hasta el final de la funcion.
	defer pg.Close()
	rdb, err := db.NewRedis(cfg.RedisAddr, cfg.RedisPass)
	if err != nil {
		log.Fatalf("No se pudo cargar redis %v", err)
	}
	defer rdb.Close()

	campusRepo := repository.NewCampusRepo(pg)
	pilotRepo := repository.NewPilotRepo(pg)

	hub := location.NewHub(rdb)

	app := fiber.New()

	app.Get("/health-check", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "online!"})
	})

	app.Get("/campus", func(c *fiber.Ctx) error {
		campuses, err := campusRepo.GetAll(c.Context())
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(campuses)
	})

	app.Get("/campus/:id", func(c *fiber.Ctx) error {
		campus, err := campusRepo.GetByID(c.Context(), c.Params("id"))
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "campus no encontrado"})
		}
		return c.JSON(campus)
	})

	app.Get("/campus/:id/pilot", func(c *fiber.Ctx) error {
		pilot, err := pilotRepo.GetByCampus(c.Context(), c.Params("id"))
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "piloto no encontrado"})
		}
		return c.JSON(pilot)
	})

	// WebSocket en servidor net/http separado
	// Fiber usa fasthttp que no soporta WebSocket hijack via adaptor
	wsMux := http.NewServeMux()
	wsMux.HandleFunc("/ws/pilot", location.HandlePilot(hub))
	go func() {
		log.Printf("WebSocket server en puerto %s", cfg.WSPort)
		if err := http.ListenAndServe(":"+cfg.WSPort, wsMux); err != nil {
			log.Fatalf("WS server error: %v", err)
		}
	}()

	log.Printf("Servidor en el puerto %s", cfg.AppPort)
	log.Fatal(app.Listen(":" + cfg.AppPort))
}
