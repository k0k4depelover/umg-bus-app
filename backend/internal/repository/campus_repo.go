// Paso no 7: Creamos el repository, para acceder a los datos.

package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/umg-bus-app/backend/internal/domain"
)

// *pgxpool -> Es una estructura (struct) que representa:
// un pool de conexiones a PostgreSQL
type CampusRepo struct {
	db *pgxpool.Pool
}

/*
NewCampusRepo -> Convención en Go:
NewX → función que crea algo

Recibe db *pgxpool.Pool -> Un pool de conexiones.const

CampusRepo{...}-> Crear una instancia del struct asignando el campo db
Devolvemos un puntero por -> &

Creamos un repositorio y le damos acceso a la base de datos
Esto es inyeccion de dependencias, en vez de que el repo cree
la base de datos nosotros se la pasamos.

Esto es el patrón -> Repository Pattern
*/
func NewCampusRepo(db *pgxpool.Pool) *CampusRepo {
	return &CampusRepo{db: db}
}

func (r *CampusRepo) DB() *pgxpool.Pool {
	return r.db
}

// Repositorio que nos permite obtener todos los campus de la base de datos
// Devuelve -> Error y un objeto de tipo Campus el cual es una estructura que creamos anteriormente
// el cual representa una tabla dentro de la base de datos.
func (r *CampusRepo) GetAll(ctx context.Context) ([]domain.Campus, error) {
	rows, err := r.db.Query(ctx, `
		SELECT campus_id, name, city,
               bound_sw_lat, bound_sw_lng, bound_ne_lat, bound_ne_lng,
               route_geojson, active, created_at
        FROM campus
        WHERE active = TRUE
        ORDER BY name
	`)

	if err != nil {
		return nil, fmt.Errorf("Campus getAll: %w", err)
	}
	defer rows.Close()

	var result []domain.Campus
	for rows.Next() {
		var c domain.Campus
		if err := rows.Scan(
			&c.CampusID, &c.Name, &c.City,
			&c.BoundSWLat, &c.BoundSWLng,
			&c.BoundNELat, &c.BoundNELng,
			&c.RouteGeoJSON, &c.Active, &c.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("campus scan: %w", err)
		}
		result = append(result, c)
	}
	return result, nil
}

func (r *CampusRepo) GetByID(ctx context.Context, id string) (*domain.Campus, error) {
	var c domain.Campus
	err := r.db.QueryRow(ctx,
		`SELECT campus_id, name, city,
               bound_sw_lat, bound_sw_lng, bound_ne_lat, bound_ne_lng,
               route_geojson, active, created_at FROM campus
							 WHERE campus_id = $1 AND active=TRUE
							 `, id).Scan(
		&c.CampusID, &c.Name, &c.City, &c.BoundSWLat,
		&c.BoundSWLng, &c.BoundNELat, &c.BoundNELng,
		&c.RouteGeoJSON, &c.Active, &c.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("campus GetByID: %w", err)
	}
	return &c, nil
}
