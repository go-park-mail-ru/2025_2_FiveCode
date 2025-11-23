package app

import (
	"backend/auth_service/internal/config"
	"backend/auth_service/internal/constants"
	"backend/auth_service/logger"
	"backend/auth_service/repository"
	"backend/auth_service/server"
	"backend/auth_service/usecase"
	"backend/pkg/interceptors"
	"backend/pkg/store"
	"errors"
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

	GRPCServer *grpc.Server
	Lis        net.Listener

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
	app.initGRPCServer()

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

	a.Logger.Info().Msg("Dependencies installed successfully")
}

func (a *App) initGRPCServer() {
	serviceName := constants.AuthServiceName
	port := a.Config.GRPCPort
	if port == 0 {
		a.Logger.Fatal().Msgf("grpc_port for service '%s' is not set in config", serviceName)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		a.Logger.Fatal().Err(err).Msgf("failed to listen on port %d", port)
	}
	a.Lis = lis
	a.closers = append(a.closers, lis)

	interceptorOpt := grpc.UnaryInterceptor(
		interceptors.LoggingInterceptor(a.Logger, logger.ToContext),
	)

	a.GRPCServer = grpc.NewServer(interceptorOpt)

	authRepo := repository.NewAuthRepository(a.Store.Redis.Client)
	authUsecase := usecase.NewAuthUsecase(authRepo, []byte(a.Config.CSRF.SecretKey))

	server.RegisterService(a.GRPCServer, authUsecase)
}

func (a *App) Run() {
	a.Logger.Info().Str("addr", a.Lis.Addr().String()).Msg("gRPC server is ready to accept connections")
	if err := a.GRPCServer.Serve(a.Lis); err != nil {
		a.Logger.Fatal().Err(err).Msg("gRPC server failed to serve")
	}
}

func (a *App) Close() error {
	a.Logger.Info().Msg("Closing application resources...")

	if a.GRPCServer != nil {
		a.GRPCServer.GracefulStop()
	}

	var errs error
	for _, closer := range a.closers {
		if err := closer.Close(); err != nil {
			a.Logger.Error().Err(err).Msg("failed to close resource")
			errs = errors.Join(errs, err)
		}
	}

	if errs != nil {
		a.Logger.Error().Err(errs).Msg("Application resources closed with errors")
	} else {
		a.Logger.Info().Msg("Application resources closed successfully")
	}

	return errs
}
