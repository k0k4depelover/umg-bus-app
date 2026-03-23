package graph

import (
	"github.com/umg-bus-app/backend/internal/location"
	"github.com/umg-bus-app/backend/internal/repository"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type Resolver struct {
	CampusRepo  *repository.CampusRepo
	PilotRepo   *repository.PilotRepo
	StudentRepo *repository.StudentRepo
	Hub         *location.Hub
}
