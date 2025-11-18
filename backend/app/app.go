package app

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"backend/config"
	"backend/store"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

const (
	PostgresDependency = "postgres"
	RedisDependency    = "redis"
	MinioDependency    = "minio"
)

type App struct {
	Config *config.Config
	Store  *store.Store

	closers []io.Closer
}

func NewApp() *App {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	})

	configPath, err := config.ReadConfigPath()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read config path from .env")
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatal().Err(err).Str("config_path", configPath).Msg("failed to load config")
	}

	return &App{
		Config:  cfg,
		Store:   store.NewStore(),
		closers: []io.Closer{},
	}
}

func (a *App) InstallDependencies(dependencies ...string) error {
	log.Info().Strs("dependencies", dependencies).Msg("Installing dependencies")

	depsMap := make(map[string]bool)
	for _, dep := range dependencies {
		depsMap[dep] = true
	}

	if depsMap[PostgresDependency] {
		log.Info().Msg("Initializing Postgres...")
		if err := a.Store.InitPostgres(a.Config); err != nil {
			return fmt.Errorf("failed to init postgres: %w", err)
		}

		if err := a.Store.Postgres.RunMigrations("./migrations"); err != nil {
			log.Fatal().Err(err).Msg("failed to run migrations")
		}
		log.Info().Msg("Migrations run successfully")

		a.closers = append(a.closers, a.Store.Postgres)
	}

	if depsMap[RedisDependency] {
		log.Info().Msg("Initializing Redis...")
		if err := a.Store.InitRedis(a.Config); err != nil {
			return fmt.Errorf("failed to init redis: %w", err)
		}
		a.closers = append(a.closers, a.Store.Redis)
	}

	if depsMap[MinioDependency] {
		log.Info().Msg("Initializing Minio...")
		if err := a.Store.InitMinioStorage(a.Config); err != nil {
			return fmt.Errorf("failed to init minio: %w", err)
		}
	}

	log.Info().Msg("Dependencies installed successfully")
	return nil
}

func (a *App) Close() error {
	log.Info().Msg("Closing application resources...")
	for _, closer := range a.closers {
		if err := closer.Close(); err != nil {
			return fmt.Errorf("failed to close resource: %w", err)
		}
	}
	log.Info().Msg("Application resources closed successfully")
	return nil
}

type RegisterGRPCServiceFunc func(srv *grpc.Server, app *App)

func (a *App) RunGRPCServer(serviceName string, registerService RegisterGRPCServiceFunc, opts ...grpc.ServerOption) {
	serviceConf, ok := a.Config.Services[serviceName]
	
	if !ok {
		log.Fatal().Msgf("config for service '%s' not found", serviceName)
	}
	port := serviceConf.GrpcPort
	if port == 0 {
		log.Fatal().Msgf("grpc_port for service '%s' is not set in config", serviceName)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal().Err(err).Msgf("failed to listen on port %d", port)
	}
	a.closers = append(a.closers, lis)

	grpcServer := grpc.NewServer(opts...)

	registerService(grpcServer, a)

	log.Info().Str("addr", lis.Addr().String()).Msg("gRPC server is ready to accept connections")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal().Err(err).Msg("gRPC server failed to serve")
	}
}

type CreateHttpHandlerFunc func(app *App) http.Handler

func (a *App) RunHTTPServer(createHandler CreateHttpHandlerFunc) {
	serverAddr, err := config.ReadServerAddress()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read server address")
	}

	handler := createHandler(a)

	server := &http.Server{
		Addr:    serverAddr,
		Handler: handler,
	}

	log.Info().Str("addr", server.Addr).Msg("HTTP server is ready to accept connections")

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal().Err(err).Msg("HTTP server failed to serve")
	}
}
