package main

import (
	"backend/internal/app"
	BlockRepo "backend/note_service/blocks/repository"
	BlockUC "backend/note_service/blocks/usecase"
	NoteRepo "backend/note_service/notes/repository"
	NoteUC "backend/note_service/notes/usecase"
	"backend/note_service/server"
	"backend/pkg/constants"
	"backend/pkg/interceptors"
	"google.golang.org/grpc"

	"github.com/rs/zerolog/log"
)

func main() {
	application := app.NewApp()
	if err := application.InstallDependencies(app.PostgresDependency); err != nil {
		log.Error().Err(err).Msg("Failed to install dependencies")
	}

	defer func() {
		if err := application.Close(); err != nil {
			log.Printf("Error closing app resources: %v", err)
		}
	}()

	registerNoteService := func(srv *grpc.Server, app *app.App) {
		notesRepo := NoteRepo.NewNotesRepository(app.Store.Postgres.DB)
		blocksRepo := BlockRepo.NewBlocksRepository(app.Store.Postgres.DB)

		notesUC := NoteUC.NewNoteUsecase(notesRepo)
		blocksUC := BlockUC.NewBlocksUsecase(blocksRepo, notesRepo)

		server.RegisterServices(srv, notesUC, blocksUC)
	}

	interceptorOpt := grpc.UnaryInterceptor(interceptors.LoggingInterceptor)
	application.RunGRPCServer(constants.NotesServiceName, registerNoteService, interceptorOpt)
}
