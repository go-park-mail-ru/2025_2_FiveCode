package app

import (
	"backend/auth_service/internal/config"
	"backend/auth_service/internal/constants"
	"backend/auth_service/logger"
	"backend/pkg/store"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type App struct {
	Config   *config.Config
	Store    *store.Store
	Logger   zerolog.Logger
	UserConn *grpc.ClientConn

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
	a.Logger.Info().Msg("Initializing dependencies for Auth Service")

	a.Logger.Info().Msg("Initializing Redis...")
	if err := a.Store.InitRedis(&store.RedisConfig{
		Host:     a.Config.Redis.Host,
		Port:     a.Config.Redis.Port,
		Password: a.Config.Redis.Password,
		DB:       a.Config.Redis.DB,
	}); err != nil {
		a.Logger.Fatal().Err(err).Msg("failed to init redis")
	}
	a.closers = append(a.closers, a.Store.Redis)

	a.Logger.Info().Msg("Connecting to User Service...")

	userHost := a.Config.Services["user"].GrpcHost
	userPort := a.Config.Services["user"].GrpcPort
	userAddr := fmt.Sprintf("%s:%d", userHost, userPort)

	conn, err := grpc.Dial(
		userAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		a.Logger.Fatal().Err(err).Msgf("failed to connect to user service at %s", userAddr)
	}

	a.UserConn = conn
	a.closers = append(a.closers, conn)

	a.Logger.Info().Msg("Dependencies installed successfully")
}

func (a *App) Close() error {
	a.Logger.Info().Msg("Closing application resources...")
	for _, closer := range a.closers {
		if err := closer.Close(); err != nil {
			a.Logger.Error().Err(err).Msg("failed to close resource")
		}
	}
	a.Logger.Info().Msg("Application resources closed")
	return nil
}

type RegisterGRPCServiceFunc func(srv *grpc.Server, app *App)

func (a *App) RunGRPCServer(registerService RegisterGRPCServiceFunc, opts ...grpc.ServerOption) {
	serviceName := constants.AuthServiceName

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
