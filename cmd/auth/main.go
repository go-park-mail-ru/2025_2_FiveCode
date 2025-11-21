package main

import (
	"backend/auth_service/app"
	"backend/auth_service/logger"
	"backend/auth_service/repository"
	"backend/auth_service/server"
	"backend/auth_service/usecase"
	"backend/pkg/interceptors"

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

	registerAuthService := func(srv *grpc.Server, app *app.App) {
		repo := repository.NewAuthRepository(app.Store.Postgres.DB, app.Store.Redis.Client)
		useCase := usecase.NewAuthUsecase(repo, []byte(app.Config.CSRF.SecretKey))

		server.RegisterService(srv, useCase)
	}

	interceptorOpt := grpc.UnaryInterceptor(
		interceptors.LoggingInterceptor(application.Logger, logger.ToContext),
	)

	application.RunGRPCServer(registerAuthService, interceptorOpt)
}
