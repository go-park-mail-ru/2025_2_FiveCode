package main

import (
	"backend/internal/app"
	"backend/pkg/constants"
	"backend/pkg/interceptors"
	"backend/user_service/repository"
	"backend/user_service/server"
	"backend/user_service/usecase"
	"github.com/rs/zerolog/log"

	"google.golang.org/grpc"
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

	registerUserService := func(srv *grpc.Server, app *app.App) {
		userRepo := repository.NewUserRepository(app.Store.Postgres.DB)
		userUsecase := usecase.NewUserUsecase(userRepo)

		server.RegisterService(srv, userUsecase)
	}

	interceptorOpt := grpc.UnaryInterceptor(interceptors.LoggingInterceptor)

	application.RunGRPCServer(constants.UserServiceName, registerUserService, interceptorOpt)
}
