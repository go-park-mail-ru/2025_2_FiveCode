package app

import (
	"backend/config"
	"backend/initialize"
	"backend/router"
	"backend/store"
	"context"
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

	configPath, err := config.ReadConfigPath()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read config path")
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatal().Err(err).Str("config_path", configPath).Msg("failed to load config")
	}

	log.Info().Int("session_duration_days", cfg.Auth.Cookie.SessionDuration).Msg("config loaded")

	if err := s.InitPostgres(cfg); err != nil {
		log.Fatal().Err(err).Str("host", cfg.Storages.Db.Host).Int("port", cfg.Storages.Db.Port).Msg("failed to init Postgres")
	}
	defer func() {
		if err := s.Postgres.Close(); err != nil {
			log.Error().Err(err).Msg("postgres close failed")
		}
	}()
	log.Info().Str("host", cfg.Storages.Db.Host).Int("port", cfg.Storages.Db.Port).Msg("Postgres initialized successfully")

	/*if err := s.Postgres.RunMigrations("./migrations"); err != nil {
		log.Fatal().Err(err).Msg("failed to run migrations")
	}
	log.Info().Msg("Migrations run successfully")*/

	if err := s.InitMinioStorage(cfg); err != nil {
		log.Fatal().Err(err).Msg("failed to init minio storage")
	}
	log.Info().Str("endpoint", cfg.Storages.Minio.Endpoint).Msg("MinIO storage initialized successfully")

	if err := s.InitRedis(cfg); err != nil {
		log.Fatal().Err(err).Msg("failed to init redis")
	}
	defer func() {
		if err := s.Redis.Close(); err != nil {
			log.Error().Err(err).Msg("redis close failed")
		}
	}()
	log.Info().Str("host", cfg.Storages.Redis.Host).Int("port", cfg.Storages.Redis.Port).Msg("Redis initialized successfully")

	if err := s.InitFillStore(context.Background()); err != nil {
		log.Fatal().Err(err).Msg("failed to fill store")
	}

	deliveries := initialize.InitDeliveries(s, cfg)

	r := router.NewRouter(s, deliveries, cfg)

	serverAddr, err := config.ReadServerAddress()
	if err != nil {
		return fmt.Errorf("failed to read server address: %w", err)
	}

	server := &http.Server{
		Addr:    serverAddr,
		Handler: r,
	}

	log.Info().Str("address", server.Addr).Msg("starting server")
	err = server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}
