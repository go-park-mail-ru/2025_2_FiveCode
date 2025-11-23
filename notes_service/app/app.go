package app

import (
	BlockRepo "backend/notes_service/blocks/repository"
	BlockUC "backend/notes_service/blocks/usecase"
	"backend/notes_service/internal/config"
	"backend/notes_service/internal/constants"
	"backend/notes_service/logger"
	NoteRepo "backend/notes_service/notes/repository"
	NoteUC "backend/notes_service/notes/usecase"
	"backend/notes_service/server"
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

	a.Logger.Info().Msg("Running migrations...")
	if err := a.Store.Postgres.RunMigrations("./db/migrations"); err != nil {
		a.Logger.Fatal().Err(err).Msg("failed to run migrations")
	}

	a.Logger.Info().Msg("Dependencies installed successfully")
}

func (a *App) initGRPCServer() {
	serviceName := constants.NotesServiceName
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

	notesRepo := NoteRepo.NewNotesRepository(a.Store.Postgres.DB)
	blocksRepo := BlockRepo.NewBlocksRepository(a.Store.Postgres.DB)

	notesUC := NoteUC.NewNoteUsecase(notesRepo)
	blocksUC := BlockUC.NewBlocksUsecase(blocksRepo, notesRepo)

	server.RegisterServices(a.GRPCServer, notesUC, blocksUC)
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
