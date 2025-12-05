package store

import "time"

type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string

	// Connection Pool настройки (опционально)
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int // в секундах
	ConnMaxIdleTime int // в секундах

	// Таймауты PostgreSQL (опционально)
	StatementTimeout int // в секундах - максимальное время выполнения запроса
	LockTimeout      int // в секундах - максимальное время ожидания блокировки
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type MinioConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Secure    bool
}

type Store struct {
	Postgres *PostgresDB
	Redis    *RedisDB
	Minio    *MinioStorage
}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) InitPostgres(config *PostgresConfig) error {
	// Проверяем, заданы ли параметры пула
	if config.MaxOpenConns > 0 {
		// Используем настройки пула из конфига
		poolConfig := ConnectionPoolConfig{
			MaxOpenConns:    config.MaxOpenConns,
			MaxIdleConns:    config.MaxIdleConns,
			ConnMaxLifetime: time.Duration(config.ConnMaxLifetime) * time.Second,
			ConnMaxIdleTime: time.Duration(config.ConnMaxIdleTime) * time.Second,

			// Таймауты
			StatementTimeout: time.Duration(config.StatementTimeout) * time.Second,
			LockTimeout:      time.Duration(config.LockTimeout) * time.Second,
		}

		db, err := NewPostgresDBWithPool(
			config.Host,
			config.Port,
			config.User,
			config.Password,
			config.DBName,
			config.SSLMode,
			poolConfig,
		)
		if err != nil {
			return err
		}
		s.Postgres = db
	} else {
		// Используем настройки по умолчанию
		poolConfig := DefaultConnectionPoolConfig()

		// Переопределяем таймауты если заданы
		if config.StatementTimeout > 0 {
			poolConfig.StatementTimeout = time.Duration(config.StatementTimeout) * time.Second
		}
		if config.LockTimeout > 0 {
			poolConfig.LockTimeout = time.Duration(config.LockTimeout) * time.Second
		}

		db, err := NewPostgresDBWithPool(
			config.Host,
			config.Port,
			config.User,
			config.Password,
			config.DBName,
			config.SSLMode,
			poolConfig,
		)
		if err != nil {
			return err
		}
		s.Postgres = db
	}

	return nil
}

func (s *Store) InitRedis(config *RedisConfig) error {
	db, err := NewRedisDB(config.Host, config.Port, config.Password, config.DB)
	if err != nil {
		return err
	}
	s.Redis = db
	return nil
}

func (s *Store) InitMinioStorage(config *MinioConfig) error {
	storage, err := NewMinioStorage(config.Endpoint, config.AccessKey, config.SecretKey, config.Secure)
	if err != nil {
		return err
	}
	s.Minio = storage
	return nil
}
