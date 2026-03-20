// Paso no 8. Lo mismo aca definimos el repositorio pero para obtener
// los valores del repositorio del piloto.

package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/umg-bus-app/backend/internal/domain"
)

type PilotRepo struct {
	db *pgxpool.Pool
}

func NewPilotRepo(db *pgxpool.Pool) *PilotRepo {
	return &PilotRepo{db: db}
}

func (r *PilotRepo) GetByCampus(ctx context.Context, campusID string) (*domain.Pilot, error) {
	var p domain.Pilot
	err := r.db.QueryRow(ctx, `
		SELECT pilot_id, campus_id, username, full_name,
               phone, active, created_at, last_seen_at
        FROM pilots
        WHERE campus_id = $1 AND active = TRUE
        LIMIT 1
	`, campusID).Scan(&p.PilotID, &p.CampusID, &p.Username,
		&p.Phone, &p.Active, &p.CreatedAt, &p.LastSeenAt,
	)
	if err != nil {
		return nil, fmt.Errorf("pilot GetByCampus: %w", err)

	}
	return &p, nil
}

func (r *PilotRepo) UpdateLastSeen(ctx context.Context, pilotID string) error {
	_, err := r.db.Exec(ctx, `
        UPDATE pilots SET last_seen_at = now()
        WHERE pilot_id = $1
    `, pilotID)
	if err != nil {
		return fmt.Errorf("pilot UpdateLastSeen: %w", err)
	}
	return nil
}
