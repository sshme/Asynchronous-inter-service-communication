package postgres

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

type Config struct {
	Host string
	Port int
	User string
	Pass string
	Name string
}

func (c *Config) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, c.Pass, c.Name)
}

func NewDb(config *Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", config.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Connected to PostgreSQL database: %s", config.Name)
	return db, nil
}
