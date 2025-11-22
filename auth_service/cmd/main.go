package main

import (
	"backend/auth_service/app"
	"backend/auth_service/internal/adapter"
	"backend/auth_service/logger"
	"backend/auth_service/repository"
	"backend/auth_service/server"
	"backend/auth_service/usecase"
	"backend/pkg/interceptors"
	userPB "backend/user_service/pkg/user/v1"

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
		authRepo := repository.NewAuthRepository(app.Store.Redis.Client)

		userClient := userPB.NewUserServiceClient(app.UserConn)

		userAdapter := adapter.NewUserGRPCAdapter(userClient)

		authUsecase := usecase.NewAuthUsecase(authRepo, userAdapter, []byte(app.Config.CSRF.SecretKey))

		server.RegisterService(srv, authUsecase)
	}

	interceptorOpt := grpc.UnaryInterceptor(
		interceptors.LoggingInterceptor(application.Logger, logger.ToContext),
	)

	application.RunGRPCServer(registerAuthService, interceptorOpt)
}
