// Paso no. 6: Definir los modelos,
// Aca representamos las tablas de la base de datos mediante
// structs de GO. Deben coincidir con los datos que tenemos en las
// tablas de nuestro motor de BD.

package domain

import (
	"time"

	"github.com/google/uuid"
)

// Nombre // Tipo // Como se va a ver en JSON

type Campus struct {
	CampusID     uuid.UUID `json:"campus_id"`
	Name         string    `json:"name"`
	City         string    `json:"city"`
	BoundSWLat   float64   `json:"bound_sw_lat"`
	BoundSWLng   float64   `json:"bound_sw_lng"`
	BoundNELat   float64   `json:"bound_ne_lat"`
	BoundNELng   float64   `json:"bound_ne_lng"`
	RouteGeoJSON *string   `json:"route_geojson,omitempty"`
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"created_at"`
}
