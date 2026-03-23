// Paso 1. (Los primeros pasos deben ser definir la base de datos,
// definir los contenedores que las manejen, variables de entorno, etc.
// )

package config

import (
	"fmt"
	"os"
)

type Config struct {
	PostgresDSN string
	RedisAddr   string
	RedisPass   string
	AppPort     string
	WSPort      string
	JWTSecret   string
}

// Load -> Nombre
// () -> Parametros de entrada
// (*Config, error) -> valores de retorno, si es solo uno, podemos devolver solo el parametro, ojo ver que la
// devolucion es solo el tipo de dato, no esta asociado a ninguna estructura.
// 	El asterisco indica que es un puntero, Config es una estructura
// error -> Manejo de errores de GO
// Es el sistema estándar de manejo de errores en Go
// Si algo falla → devuelve un error
// Si todo va bien → devuelve nil
// := -> Significa que es declaracion + asignacion
// = -> Asignacion, solo cuando la variable ya existe.

func Load() (*Config, error) {
	pgHost := getEnv("POSTGRES_HOST", "localhost")
	pgPort := getEnv("POSTGRES_PORT", "5436")
	pgUser := getEnv("POSTGRES_USER", "admin")
	pgPass := getEnv("POSTGRES_PASSWORD", "postgresumgtracker")
	pgDB := getEnv("POSTGRES_DB", "transportation-tracker-db")
	JWTSecret := getEnv("JWT_SECRET", "dev-secret-cambia-en-produccion")

	// & -> Es la direccion de memoria de una variable
	// URL format requerido por golang-migrate; pgxpool también lo acepta.
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		pgUser, pgPass, pgHost, pgPort, pgDB,
	)
	return &Config{
		PostgresDSN: dsn,
		RedisAddr:   getEnv("REDIS_ADDR", "localhost:6381"),
		RedisPass:   getEnv("REDIS_PASS", ""),
		AppPort:     getEnv("APP_PORT", "8084"),
		WSPort:      getEnv("WS_PORT", "8085"),
		JWTSecret:   JWTSecret,
	}, nil
}

// Estructura de un IF:
// if <declaración>; <condición>
// Declaramos la variable como la obtencion de la funcion Getenv
// Si no es "", retorna el valor
func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback // retorna ""
}
