package main

import (
	"backend/gateway_service/app"

	authDelivery "backend/gateway_service/internal/auth/delivery"
	notesDelivery "backend/gateway_service/internal/notes/delivery"
	userDelivery "backend/gateway_service/internal/user/delivery"

	fileDelivery "backend/gateway_service/internal/file/delivery"

	"backend/gateway_service/router"

	authPB "backend/auth_service/pkg/auth/v1"
	blockPB "backend/notes_service/pkg/block/v1"
	notePB "backend/notes_service/pkg/note/v1"
	userPB "backend/user_service/pkg/user/v1"

	authRepo "backend/gateway_service/internal/auth/repository"
	fileRepo "backend/gateway_service/internal/file/repository"
	notesRepo "backend/gateway_service/internal/notes/repository"
	userRepo "backend/gateway_service/internal/user/repository"

	authUC "backend/gateway_service/internal/auth/usecase"
	fileUC "backend/gateway_service/internal/file/usecase"
	notesUC "backend/gateway_service/internal/notes/usecase"
	userUC "backend/gateway_service/internal/user/usecase"

	"time"

	"github.com/rs/zerolog/log"
)

func main() {
	application := app.NewApp()
	defer func() {
		if err := application.Close(); err != nil {
			log.Error().Err(err).Msg("Error closing resources")
		}
	}()

	authClientGRPC := authPB.NewAuthClient(application.AuthConn)
	userClientGRPC := userPB.NewUserServiceClient(application.UserConn)
	noteClientGRPC := notePB.NewNoteServiceClient(application.NotesConn)
	blockClientGRPC := blockPB.NewBlockServiceClient(application.NotesConn)

	gatewayAuthRepo := authRepo.NewAuthRepository(authClientGRPC)
	gatewayUserRepo := userRepo.NewUserRepository(userClientGRPC)
	gatewayNotesRepo := notesRepo.NewNotesRepository(noteClientGRPC, blockClientGRPC)

	gatewayFileRepo := fileRepo.NewFileRepository(application.Store.Postgres.DB, application.Store.Minio.Client)

	gatewayAuthUC := authUC.NewAuthUsecase(gatewayAuthRepo, gatewayUserRepo)
	gatewayUserUC := userUC.NewUserUsecase(gatewayUserRepo, gatewayAuthRepo)
	gatewayNotesUC := notesUC.NewNotesUsecase(gatewayNotesRepo)
	gatewayFileUC := fileUC.NewFileUsecase(gatewayFileRepo)

	sessionDuration := time.Duration(application.Config.Cookie.SessionDuration) * time.Hour

	authHandler := authDelivery.NewAuthDelivery(gatewayAuthUC, sessionDuration)
	userHandler := userDelivery.NewUserDelivery(gatewayUserUC)
	notesHandler := notesDelivery.NewNotesDelivery(gatewayNotesUC)
	fileHandler := fileDelivery.NewFileDelivery(gatewayFileUC)

	sessionValidator := gatewayAuthRepo

	mainRouter := router.NewRouter(
		application.Config,
		&application.Logger,
		sessionValidator,
		authHandler,
		userHandler,
		notesHandler,
		fileHandler,
	)

	application.RunHTTPServer(mainRouter)
}
