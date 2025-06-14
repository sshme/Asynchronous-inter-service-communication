package postgres

import (
	"database/sql"
	"embed"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func RunMigrations(db *sql.DB) error {
	log.Println("Starting migrations...")

	entries, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("could not read migrations directory: %w", err)
	}

	log.Printf("Found %d migration files:", len(entries))
	for _, entry := range entries {
		log.Printf("  - %s", entry.Name())
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create postgres driver: %w", err)
	}

	d, err := iofs.New(migrationFiles, "migrations")
	if err != nil {
		return fmt.Errorf("could not create iofs source: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", d, "postgres", driver)
	if err != nil {
		return fmt.Errorf("could not create migrate instance: %w", err)
	}

	log.Println("Running migrations...")
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("could not run migrations: %w", err)
	}

	log.Println("Migrations completed successfully")
	return nil
}

func MigrateDown(db *sql.DB, steps int) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create postgres driver: %w", err)
	}

	d, err := iofs.New(migrationFiles, "migrations")
	if err != nil {
		return fmt.Errorf("could not create iofs source: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", d, "postgres", driver)
	if err != nil {
		return fmt.Errorf("could not create migrate instance: %w", err)
	}

	if err := m.Steps(-steps); err != nil {
		return fmt.Errorf("could not run down migrations: %w", err)
	}

	return nil
}
