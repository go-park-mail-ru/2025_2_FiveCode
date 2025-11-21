package main

import (
	"backend/notes_service/logger"
	"backend/notes_service/app"
	BlockRepo "backend/notes_service/blocks/repository"
	BlockUC "backend/notes_service/blocks/usecase"
	NoteRepo "backend/notes_service/notes/repository"
	NoteUC "backend/notes_service/notes/usecase"
	"backend/notes_service/server"
	"backend/pkg/interceptors"

	"google.golang.org/grpc"

	"github.com/rs/zerolog/log"
)

func main() {
	application := app.NewApp()

	defer func() {
		if err := application.Close(); err != nil {
			log.Error().Err(err).Msgf("Error closing app resources: %v", err)
		}
	}()

	registerNoteService := func(srv *grpc.Server, app *app.App) {
		notesRepo := NoteRepo.NewNotesRepository(app.Store.Postgres.DB)
		blocksRepo := BlockRepo.NewBlocksRepository(app.Store.Postgres.DB)

		notesUC := NoteUC.NewNoteUsecase(notesRepo)
		blocksUC := BlockUC.NewBlocksUsecase(blocksRepo, notesRepo)

		server.RegisterServices(srv, notesUC, blocksUC)
	}

	interceptorOpt := grpc.UnaryInterceptor(
		interceptors.LoggingInterceptor(application.Logger, logger.ToContext),
	)
	
	application.RunGRPCServer(registerNoteService, interceptorOpt)
}
