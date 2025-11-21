package app

import (
	"backend/notes_service/internal/config"
	"backend/notes_service/internal/constants"
	"backend/notes_service/logger"
	"backend/pkg/store"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
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
	a.Logger.Info().Msg("Initializing dependencies for Note Service")

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

type RegisterGRPCServiceFunc func(srv *grpc.Server, app *App)

func (a *App) RunGRPCServer(registerService RegisterGRPCServiceFunc, opts ...grpc.ServerOption) {
	serviceName := constants.NotesServiceName

	port := a.Config.GRPCPort
	if port == 0 {
		a.Logger.Fatal().Msgf("grpc_port for service '%s' is not set in config", serviceName)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		a.Logger.Fatal().Err(err).Msgf("failed to listen on port %d", port)
	}
	a.closers = append(a.closers, lis)

	grpcServer := grpc.NewServer(opts...)

	registerService(grpcServer, a)

	a.Logger.Info().Str("addr", lis.Addr().String()).Msg("gRPC server is ready to accept connections")
	if err := grpcServer.Serve(lis); err != nil {
		a.Logger.Fatal().Err(err).Msg("gRPC server failed to serve")
	}
}
