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
type Pilot struct {
	PilotID    uuid.UUID  `json:"pilot_id"`
	CampusID   uuid.UUID  `json:"campus_id"`
	Username   string     `json:"username"`
	FullName   string     `json:"full_name"`
	Phone      *string    `json:"phone,omitempty"`
	Active     bool       `json:"active"`
	CreatedAt  time.Time  `json:"created_at"`
	LastSeenAt *time.Time `json:"last_seen_at,omitempty"`
	// password_hash y secret_code nunca van en el struct de respuesta
}
