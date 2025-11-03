package store

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/lib/pq"
)

type PostgresDB struct {
	DB *sql.DB
}

func NewPostgresDB(host string, port int, user, password, dbname, sslmode string) (*PostgresDB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, errors.New("Error connecting to Postgres database: " + err.Error())
	}

	if err := db.Ping(); err != nil {
		return nil, errors.New("failed to ping database: " + err.Error())
	}

	return &PostgresDB{DB: db}, nil
}

func (p *PostgresDB) Close() error {
	return p.DB.Close()
}

func (p *PostgresDB) RunMigrations(migrationsPath string) error {
	driver, err := postgres.WithInstance(p.DB, &postgres.Config{})
	if err != nil {
		return errors.New("failed to create migrate driver: " + err.Error())
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres",
		driver,
	)
	if err != nil {
		return errors.New("failed to create migrate instance: " + err.Error())
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return errors.New("failed to run migrations: " + err.Error())
	}

	return nil
}
