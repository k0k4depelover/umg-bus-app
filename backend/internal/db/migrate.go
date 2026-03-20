package db

import (
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(dsn string) error {
	m, err := migrate.New("file://./migrations", dsn)
	if err != nil {
		return fmt.Errorf("No fue posible encontrar el archivo de migraciones %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("Sin cambios pendientes en migrations.")
			return nil
		}
		return fmt.Errorf("Error corriendo migrations. %w", err)

	}

	log.Println("Migrations aplicadas")
	return nil
}
