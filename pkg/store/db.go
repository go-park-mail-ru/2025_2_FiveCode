package store

import (
	"backend/pkg/metrics"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

type Tx interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	Commit() error
	Rollback() error
}

type DB interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error)
	Close() error
	GetSQLDB() *sql.DB
}

type dbWrapper struct {
	*sql.DB
}

func (d *dbWrapper) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	start := time.Now()
	row := d.DB.QueryRowContext(ctx, query, args...)
	metrics.RecordDBQueryDuration(start, "query", "db")
	return row
}

func (d *dbWrapper) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := d.DB.ExecContext(ctx, query, args...)
	if err != nil {
		metrics.RecordDBQueryError("exec", "db")
	} else {
		metrics.RecordDBQueryDuration(start, "exec", "db")
	}
	return result, err
}

func (d *dbWrapper) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := d.DB.QueryContext(ctx, query, args...)
	if err != nil {
		metrics.RecordDBQueryError("query", "db")
	} else {
		metrics.RecordDBQueryDuration(start, "query", "db")
	}
	return rows, err
}

func (d *dbWrapper) BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error) {
	tx, err := d.DB.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &txWrapper{Tx: tx}, nil
}

func (d *dbWrapper) GetSQLDB() *sql.DB {
	return d.DB
}

type txWrapper struct {
	*sql.Tx
}

func (t *txWrapper) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	start := time.Now()
	row := t.Tx.QueryRowContext(ctx, query, args...)
	metrics.RecordDBQueryDuration(start, "query", "db")
	return row
}

func (t *txWrapper) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := t.Tx.ExecContext(ctx, query, args...)
	if err != nil {
		metrics.RecordDBQueryError("exec", "db")
	} else {
		metrics.RecordDBQueryDuration(start, "exec", "db")
	}
	return result, err
}

func (t *txWrapper) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := t.Tx.QueryContext(ctx, query, args...)
	if err != nil {
		metrics.RecordDBQueryError("query", "db")
	} else {
		metrics.RecordDBQueryDuration(start, "query", "db")
	}
	return rows, err
}

func (t *txWrapper) Commit() error {
	return t.Tx.Commit()
}

func (t *txWrapper) Rollback() error {
	return t.Tx.Rollback()
}

type PostgresDB struct {
	DB DB
}

func NewPostgresDB(host string, port int, user, password, dbname, sslmode string) (*PostgresDB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("error connecting to Postgres database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresDB{DB: &dbWrapper{DB: db}}, nil
}

func (p *PostgresDB) Close() error {
	return p.DB.Close()
}

func (p *PostgresDB) RunMigrations(migrationsPath string) error {
	driver, err := postgres.WithInstance(p.DB.GetSQLDB(), &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migrate driver: %w", err)
	}

	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to resolve migration path: %w", err)
	}

	migrationURL := fmt.Sprintf("file://%s", absPath)
	m, err := migrate.NewWithDatabaseInstance(
		migrationURL,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance (path: %s): %w", migrationURL, err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
