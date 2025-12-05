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

// ============================================
// Connection Pool Configuration
// ============================================

// ConnectionPoolConfig содержит настройки пула соединений
type ConnectionPoolConfig struct {
	MaxOpenConns    int           // Максимальное количество открытых соединений
	MaxIdleConns    int           // Максимальное количество простаивающих соединений
	ConnMaxLifetime time.Duration // Максимальное время жизни соединения
	ConnMaxIdleTime time.Duration // Максимальное время простоя соединения

	// Таймауты PostgreSQL (устанавливаются для каждой сессии)
	StatementTimeout time.Duration // Максимальное время выполнения запроса
	LockTimeout      time.Duration // Максимальное время ожидания блокировки
}

// DefaultConnectionPoolConfig возвращает рекомендуемые настройки для большинства случаев
func DefaultConnectionPoolConfig() ConnectionPoolConfig {
	return ConnectionPoolConfig{
		MaxOpenConns:    25,              // Достаточно для большинства микросервисов
		MaxIdleConns:    5,               // ~20% от MaxOpenConns
		ConnMaxLifetime: 5 * time.Minute, // Предотвращает использование старых соединений
		ConnMaxIdleTime: 5 * time.Minute, // Закрываем простаивающие соединения

		// Таймауты по умолчанию (консервативные значения)
		StatementTimeout: 30 * time.Second, // 30 секунд на запрос
		LockTimeout:      10 * time.Second, // 10 секунд на получение блокировки
	}
}

type PostgresDB struct {
	DB DB
}

// NewPostgresDB создает новое подключение к PostgreSQL без настройки пула
// Используйте NewPostgresDBWithPool для настройки пула соединений
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

// NewPostgresDBWithPool создает новое подключение к PostgreSQL с настройками пула соединений
func NewPostgresDBWithPool(host string, port int, user, password, dbname, sslmode string, poolConfig ConnectionPoolConfig) (*PostgresDB, error) {
	// ============================================
	// Строим DSN с параметрами таймаутов
	// ============================================
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	// Добавляем options для установки таймаутов через DSN
	// Это применится ко всем соединениям в пуле автоматически
	if poolConfig.StatementTimeout > 0 || poolConfig.LockTimeout > 0 {
		options := ""

		if poolConfig.StatementTimeout > 0 {
			timeoutMs := int(poolConfig.StatementTimeout.Milliseconds())
			options += fmt.Sprintf("-c statement_timeout=%d ", timeoutMs)
		}

		if poolConfig.LockTimeout > 0 {
			timeoutMs := int(poolConfig.LockTimeout.Milliseconds())
			options += fmt.Sprintf("-c lock_timeout=%d", timeoutMs)
		}

		if options != "" {
			dsn += fmt.Sprintf(" options='%s'", options)
		}
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("error connecting to Postgres database: %w", err)
	}

	// ============================================
	// Настройка Connection Pool
	// ============================================

	// MaxOpenConns: максимальное количество одновременно открытых соединений
	// Включает и используемые, и простаивающие соединения
	db.SetMaxOpenConns(poolConfig.MaxOpenConns)

	// MaxIdleConns: максимальное количество простаивающих соединений
	// Рекомендуется 20-30% от MaxOpenConns
	db.SetMaxIdleConns(poolConfig.MaxIdleConns)

	// ConnMaxLifetime: максимальное время жизни соединения
	// Предотвращает использование "старых" соединений
	// Полезно если БД принудительно закрывает долгоживущие соединения
	db.SetConnMaxLifetime(poolConfig.ConnMaxLifetime)

	// ConnMaxIdleTime: максимальное время, которое соединение может простаивать
	// Закрывает неиспользуемые соединения для экономии ресурсов
	db.SetConnMaxIdleTime(poolConfig.ConnMaxIdleTime)

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
