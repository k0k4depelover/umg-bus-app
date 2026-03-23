package domain

import (
	"time"

	"github.com/google/uuid"
)

type Student struct {
	StudentID uuid.UUID `json:"student_id"`
	CampusID  uuid.UUID `json:"campus_id"`
	Username  string    `json:"username"`
	FullName  string    `json:"full_name"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
}
