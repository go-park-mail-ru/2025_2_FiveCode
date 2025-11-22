package main

import (
	"backend/gateway_service/app"
	"backend/gateway_service/internal/adapter"

	authDelivery "backend/gateway_service/internal/delivery/auth"
	notesDelivery "backend/gateway_service/internal/delivery/notes"
	userDelivery "backend/gateway_service/internal/delivery/user"

	fileDelivery "backend/gateway_service/file/delivery"
	fileRepository "backend/gateway_service/file/repository"
	fileUsecase "backend/gateway_service/file/usecase"

	"backend/gateway_service/router"

	authPB "backend/auth_service/pkg/auth/v1"
	blockPB "backend/notes_service/pkg/block/v1"
	notePB "backend/notes_service/pkg/note/v1"
	userPB "backend/user_service/pkg/user/v1"

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

	authClient := authPB.NewAuthClient(application.AuthConn)
	userClient := userPB.NewUserServiceClient(application.UserConn)
	noteClient := notePB.NewNoteServiceClient(application.NotesConn)
	blockClient := blockPB.NewBlockServiceClient(application.NotesConn)

	sessionValidator := adapter.NewAuthGrpcAdapter(authClient)

	sessionDuration := time.Duration(application.Config.Cookie.SessionDuration) * time.Hour
	authHandler := authDelivery.NewAuthDelivery(authClient, userClient, sessionDuration)

	userHandler := userDelivery.NewUserDelivery(userClient, authClient)

	notesHandler := notesDelivery.NewNotesDelivery(noteClient)
	blocksHandler := notesDelivery.NewBlocksDelivery(blockClient)

	fileRepo := fileRepository.NewFileRepository(application.Store.Postgres.DB, application.Store.Minio.Client)
	fileUC := fileUsecase.NewFileUsecase(fileRepo)
	fileHandler := fileDelivery.NewFileDelivery(fileUC)

	mainRouter := router.NewRouter(
		application.Config,
		&application.Logger,
		sessionValidator,
		authHandler,
		userHandler,
		notesHandler,
		blocksHandler,
		fileHandler,
	)

	application.RunHTTPServer(mainRouter)
}
