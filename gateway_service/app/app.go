package app

import (
	"backend/gateway_service/internal/config"
	"backend/gateway_service/logger"
	"backend/pkg/store"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type App struct {
	Config *config.Config
	Store  *store.Store
	Logger zerolog.Logger

	closers []io.Closer
}

func NewApp() *App {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	})

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	appLogger := logger.New()

	app := &App{
		Config:  cfg,
		Store:   store.NewStore(),
		Logger:  appLogger,
		closers: []io.Closer{},
	}

	app.initDependencies()

	return app

}

func (a *App) initDependencies() {
	a.Logger.Info().Msg("Initializing dependencies for Gateway Service")

	a.Logger.Info().Msg("Initializing Postgres...")
	if err := a.Store.InitPostgres(&store.PostgresConfig{
		Host:     a.Config.DB.Host,
		Port:     a.Config.DB.Port,
		User:     a.Config.DB.User,
		Password: a.Config.DB.Password,
		DBName:   a.Config.DB.DBName,
		SSLMode:  a.Config.DB.SSLMode,
	}); err != nil {
		a.Logger.Fatal().Err(err).Msg("failed to init postgres")
	}
	a.closers = append(a.closers, a.Store.Postgres)

	a.Logger.Info().Msg("Initializing Minio...")
	if err := a.Store.InitMinioStorage(&store.MinioConfig{
		Endpoint:  a.Config.Minio.Endpoint,
		AccessKey: a.Config.Minio.AccessKey,
		SecretKey: a.Config.Minio.SecretKey,
		Secure:    a.Config.Minio.Secure,
	}); err != nil {
		a.Logger.Fatal().Err(err).Msg("failed to init minio")
	}

	a.Logger.Info().Msg("Dependencies installed successfully")

}

func (a *App) Close() error {
	a.Logger.Info().Msg("Closing application resources...")
	for _, closer := range a.closers {
		if err := closer.Close(); err != nil {
			return fmt.Errorf("failed to close resource: %w", err)
		}
	}
	a.Logger.Info().Msg("Application resources closed successfully")
	return nil
}

type CreateHttpHandlerFunc func(app *App) http.Handler

func (a *App) RunHTTPServer(createHandler CreateHttpHandlerFunc) {
	serverAddr := a.Config.Server.Host + ":" + fmt.Sprint(a.Config.Server.Port)

	handler := createHandler(a)

	server := &http.Server{
		Addr:    serverAddr,
		Handler: handler,
	}

	a.Logger.Info().Str("addr", server.Addr).Msg("HTTP server is ready to accept connections")

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		a.Logger.Fatal().Err(err).Msg("HTTP server failed to serve")
	}

}
