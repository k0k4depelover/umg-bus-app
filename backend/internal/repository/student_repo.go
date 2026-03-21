package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type StudentRepo struct {
	db *pgxpool.Pool
}

func NewStudentRepo(db *pgxpool.Pool) *StudentRepo {
	return &StudentRepo{db: db}
}

// ChangeCampus actualiza el campus del estudiante y registra el cambio en student_campus_changes.
func (r *StudentRepo) ChangeCampus(ctx context.Context, studentID string, toCampusID string) error {
	var fromCampusID string
	err := r.db.QueryRow(ctx,
		`SELECT campus_id FROM students WHERE student_id= $1`, studentID,
	).Scan(&fromCampusID)
	if err != nil {
		return fmt.Errorf("estudiante no encontrado")
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`UPDATE students SET campus_id= $1 WHERE student_id = $2`,
		toCampusID, studentID,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO student_campus_changes (student_id, from_campus_id, to_campus_id)
		VALUES($1, $2, $3)
		`, studentID, fromCampusID, toCampusID)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}
