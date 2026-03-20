// Paso 2. Definir la conexion con la base de datos.

package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPostgresPool creates a connection pool to PostgreSQL.
// Uses pgxpool for concurrent access across goroutines.
// Fmt es un paquete para imprimir, formatear strings, crear errores.
// Un error en GO es una interfaz, si algo falla el error tiene un valor,
// Ejemplo de la interfaz:
/*
type error interface {
  Error() string }
*/
// si no hay error el error es nil.
// nil es la ausencia de valor
func NewPostgres(dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres: failed to create pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("postgres: ping failed: %w", err)
	}

	return pool, nil
}
