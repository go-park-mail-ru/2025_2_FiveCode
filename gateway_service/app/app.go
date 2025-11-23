package app

import (
	"backend/gateway_service/internal/config"
	"backend/gateway_service/logger"
	"backend/gateway_service/router"
	"backend/pkg/store"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	authDelivery "backend/gateway_service/internal/auth/delivery"
	authRepo "backend/gateway_service/internal/auth/repository"
	authUC "backend/gateway_service/internal/auth/usecase"

	fileDelivery "backend/gateway_service/internal/file/delivery"
	fileRepo "backend/gateway_service/internal/file/repository"
	fileUC "backend/gateway_service/internal/file/usecase"

	notesDelivery "backend/gateway_service/internal/notes/delivery"
	notesRepo "backend/gateway_service/internal/notes/repository"
	notesUC "backend/gateway_service/internal/notes/usecase"

	userDelivery "backend/gateway_service/internal/user/delivery"
	userRepo "backend/gateway_service/internal/user/repository"
	userUC "backend/gateway_service/internal/user/usecase"

	authPB "backend/auth_service/pkg/auth/v1"
	blockPB "backend/notes_service/pkg/block/v1"
	notePB "backend/notes_service/pkg/note/v1"
	userPB "backend/user_service/pkg/user/v1"

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

	Handler http.Handler

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
	app.initHTTPHandler()

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

func (a *App) initHTTPHandler() {
	a.Logger.Info().Msg("Initializing HTTP Handlers...")

	// gRPC Clients
	authClientGRPC := authPB.NewAuthClient(a.AuthConn)
	userClientGRPC := userPB.NewUserServiceClient(a.UserConn)
	noteClientGRPC := notePB.NewNoteServiceClient(a.NotesConn)
	blockClientGRPC := blockPB.NewBlockServiceClient(a.NotesConn)

	// Repositories
	gatewayAuthRepo := authRepo.NewAuthRepository(authClientGRPC)
	gatewayUserRepo := userRepo.NewUserRepository(userClientGRPC)
	gatewayNotesRepo := notesRepo.NewNotesRepository(noteClientGRPC, blockClientGRPC)
	gatewayFileRepo := fileRepo.NewFileRepository(a.Store.Postgres.DB, a.Store.Minio.Client)

	// Usecases
	gatewayAuthUC := authUC.NewAuthUsecase(gatewayAuthRepo, gatewayUserRepo)
	gatewayUserUC := userUC.NewUserUsecase(gatewayUserRepo, gatewayAuthRepo)
	gatewayNotesUC := notesUC.NewNotesUsecase(gatewayNotesRepo)
	gatewayFileUC := fileUC.NewFileUsecase(gatewayFileRepo)

	// Handlers
	sessionDuration := time.Duration(a.Config.Cookie.SessionDuration) * time.Hour

	authHandler := authDelivery.NewAuthDelivery(gatewayAuthUC, sessionDuration)
	userHandler := userDelivery.NewUserDelivery(gatewayUserUC)
	notesHandler := notesDelivery.NewNotesDelivery(gatewayNotesUC)
	fileHandler := fileDelivery.NewFileDelivery(gatewayFileUC)

	sessionValidator := gatewayAuthRepo

	// Router
	a.Handler = router.NewRouter(
		a.Config,
		&a.Logger,
		sessionValidator,
		authHandler,
		userHandler,
		notesHandler,
		fileHandler,
	)
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
	a.Logger.Info().Msg("Closing application resources...")
	
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

func (a *App) Run() {
	addr := fmt.Sprintf("%s:%d", a.Config.Server.Host, a.Config.Server.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: a.Handler,
	}

	a.Logger.Info().Str("addr", addr).Msg("Gateway is running")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		a.Logger.Fatal().Err(err).Msg("server failed")
	}
}
