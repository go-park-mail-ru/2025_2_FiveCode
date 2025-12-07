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
	ShareRepo "backend/notes_service/sharing/repository"
	ShareUC "backend/notes_service/sharing/usecase"
	"backend/pkg/interceptors"
	"backend/pkg/metrics"
	"backend/pkg/store"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	ctx    context.Context
	cancel context.CancelFunc

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

	ctx, cancel := context.WithCancel(context.Background())

	app := &App{
		Config:  cfg,
		Store:   store.NewStore(),
		Logger:  appLogger,
		ctx:     ctx,
		cancel:  cancel,
		closers: []io.Closer{},
	}

	app.initDependencies()
	app.initGRPCServer()
	app.initMetrics()
	app.startSearchIndexRefresher()

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

	interceptorOpt := grpc.ChainUnaryInterceptor(
		interceptors.LoggingInterceptor(a.Logger, logger.ToContext),
		interceptors.MetricsInterceptor(constants.NotesServiceName),
	)

	a.GRPCServer = grpc.NewServer(interceptorOpt)

	notesRepo := NoteRepo.NewNotesRepository(a.Store.Postgres.DB)
	blocksRepo := BlockRepo.NewBlocksRepository(a.Store.Postgres.DB)
	shareRepo := ShareRepo.NewSharingRepository(a.Store.Postgres.DB)

	notesUC := NoteUC.NewNoteUsecase(notesRepo, shareRepo)
	blocksUC := BlockUC.NewBlocksUsecase(blocksRepo, notesRepo, shareRepo)
	shareUC := ShareUC.NewSharingUsecase(shareRepo, notesRepo)

	server.RegisterServices(a.GRPCServer, notesUC, blocksUC, shareUC)
}

func (a *App) initMetrics() {
	a.Logger.Info().Msg("Starting metrics server...")

	metricsPort := a.Config.MetricsPort
	if metricsPort == 0 {
		a.Logger.Fatal().Msg("metrics_port is not set in config")
	}

	go func() {
		metricsMux := http.NewServeMux()
		metricsMux.Handle("/metrics", promhttp.HandlerFor(metrics.Registry(), promhttp.HandlerOpts{}))

		metricsAddr := fmt.Sprintf(":%d", metricsPort)
		a.Logger.Info().Str("addr", metricsAddr).Msg("Metrics server is running")
		if err := http.ListenAndServe(metricsAddr, metricsMux); err != nil {
			a.Logger.Error().Err(err).Msg("metrics server failed")
		}
	}()
}

func (a *App) startSearchIndexRefresher() {
	a.Logger.Info().Msg("Starting search index refresher worker...")

	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		a.Config.DB.User,
		a.Config.DB.Password,
		a.Config.DB.Host,
		a.Config.DB.Port,
		a.Config.DB.DBName,
		a.Config.DB.SSLMode,
	)

	notesRepo := NoteRepo.NewNotesRepository(a.Store.Postgres.DB)

	ctx := logger.ToContext(a.ctx, a.Logger)
	if err := notesRepo.StartSearchIndexRefresher(ctx, connString); err != nil {
		a.Logger.Fatal().Err(err).Msg("failed to start search index refresher")
	}

	a.Logger.Info().Msg("Search index refresher worker started successfully")
}

func (a *App) Run() {
	a.Logger.Info().Str("addr", a.Lis.Addr().String()).Msg("gRPC server is ready to accept connections")
	if err := a.GRPCServer.Serve(a.Lis); err != nil {
		a.Logger.Fatal().Err(err).Msg("gRPC server failed to serve")
	}
}

func (a *App) Close() error {
	a.Logger.Info().Msg("Closing application resources...")

	if a.cancel != nil {
		a.cancel()
		a.Logger.Info().Msg("Cancelled background workers")
	}

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
