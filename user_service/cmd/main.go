package main

import (
	"backend/user_service/logger"
	"backend/pkg/interceptors"
	"backend/user_service/app"
	"backend/user_service/repository"
	"backend/user_service/server"
	"backend/user_service/usecase"

	"github.com/rs/zerolog/log"

	"google.golang.org/grpc"
)

func main() {
	application := app.NewApp()

	defer func() {
		if err := application.Close(); err != nil {
			log.Error().Err(err).Msgf("Error closing app resources: %v", err)
		}
	}()

	registerUserService := func(srv *grpc.Server, app *app.App) {
		userRepo := repository.NewUserRepository(app.Store.Postgres.DB)
		userUsecase := usecase.NewUserUsecase(userRepo)

		server.RegisterService(srv, userUsecase)
	}

	interceptorOpt := grpc.UnaryInterceptor(
		interceptors.LoggingInterceptor(application.Logger, logger.ToContext),
	)

	application.RunGRPCServer(registerUserService, interceptorOpt)
}
