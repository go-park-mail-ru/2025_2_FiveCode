package app

import (
	"backend/config"
	"backend/initialize"
	"backend/router"
	"backend/store"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
		NoColor:    false,
	})
}

func RunApp() error {
	s := store.NewStore()

	if err := s.InitFillStore(); err != nil {
		return fmt.Errorf("failed to fill store: %w", err)
	}

	configPath, err := config.ReadConfigPath()
	if err != nil {
		return fmt.Errorf("failed to read cfgig path: %w", err)
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load cfgig: %w", err)
	}

	log.Debug().
		Str("host", cfg.Storages.Db.Host).
		Int("port", cfg.Storages.Db.Port).
		Str("user", cfg.Storages.Db.User).
		Str("dbname", cfg.Storages.Db.DBName).
		Str("sslmode", cfg.Storages.Db.SSLMode).
		Msg("Attempting to connect to PostgreSQL...")

	if err := s.InitPostgres(cfg); err != nil {
		log.Fatal().
			Err(err).
			Str("host", cfg.Storages.Db.Host).
			Int("port", cfg.Storages.Db.Port).
			Msg("Failed to init PostgreSQL")
	}
	defer func() {
		if err := s.Db.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close database connection")
		}
	}()
	log.Info().Str("addr", cfg.Storages.Minio.Endpoint).Msg("Postgres initialized successfully")

	if err := s.Db.RunMigrations("./migrations"); err != nil {
		log.Fatal().Err(err).Msg("Failed to run migrations")
	}
	log.Info().Str("addr", cfg.Storages.Minio.Endpoint).Msg("Migrations run successfully")

	if err := s.InitMinioStorage(cfg); err != nil {
		return fmt.Errorf("failed to initialize minio storage: %w", err)
	}
	log.Info().Str("addr", cfg.Storages.Minio.Endpoint).Msg("MinIO storage initialized successfully")

	deliveries := initialize.InitDeliveries(s, cfg)

	r := router.NewRouter(s, deliveries)

	serverAddr, err := config.ReadServerAddress()
	if err != nil {
		return fmt.Errorf("failed to read server address: %w", err)
	}

	server := &http.Server{
		Addr:    serverAddr,
		Handler: r,
	}

	log.Info().Str("addr", server.Addr).Msg("listening")
	err = server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}
