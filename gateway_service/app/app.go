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
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type App struct {
	Config *config.Config
	Store  *store.Store
	Logger zerolog.Logger

	AuthConn  *grpc.ClientConn
	UserConn  *grpc.ClientConn
	NotesConn *grpc.ClientConn

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
	a.Logger.Info().Msg("Initializing dependencies...")

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

	a.Logger.Info().Msg("Running migrations...")
	if err := a.Store.Postgres.RunMigrations("./db/migrations"); err != nil {
		a.Logger.Fatal().Err(err).Msg("failed to run migrations")
	}

	if err := a.Store.InitMinioStorage(&store.MinioConfig{
		Endpoint:  a.Config.Minio.Endpoint,
		AccessKey: a.Config.Minio.AccessKey,
		SecretKey: a.Config.Minio.SecretKey,
		Secure:    a.Config.Minio.Secure,
	}); err != nil {
		a.Logger.Fatal().Err(err).Msg("failed to init minio")
	}

	a.AuthConn = a.mustConnectGrpc("auth")
	a.UserConn = a.mustConnectGrpc("user")
	a.NotesConn = a.mustConnectGrpc("notes")
}

func (a *App) mustConnectGrpc(serviceName string) *grpc.ClientConn {
	cfg, ok := a.Config.Services[serviceName]
	if !ok {
		a.Logger.Fatal().Msgf("config for service '%s' not found", serviceName)
	}
	addr := fmt.Sprintf("%s:%d", cfg.GrpcHost, cfg.GrpcPort)

	a.Logger.Info().Str("service", serviceName).Str("addr", addr).Msg("connecting to gRPC service")

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		a.Logger.Fatal().Err(err).Msgf("failed to connect to %s", serviceName)
	}

	a.closers = append(a.closers, conn)
	return conn
}

func (a *App) Close() error {
	for _, closer := range a.closers {
		_ = closer.Close()
	}
	return nil
}

func (a *App) RunHTTPServer(handler http.Handler) {
	addr := fmt.Sprintf("%s:%d", a.Config.Server.Host, a.Config.Server.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	a.Logger.Info().Str("addr", addr).Msg("Gateway is running")
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		a.Logger.Fatal().Err(err).Msg("server failed")
	}
}
