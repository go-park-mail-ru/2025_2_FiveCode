package main

import (
	"backend/internal/facade/app"
	"backend/internal/user_service/repository"
	"backend/internal/user_service/server"
	"backend/internal/user_service/usecase"
	"backend/pkg/constants"
	"backend/pkg/interceptors"
	"log"

	"google.golang.org/grpc"
)

func main() {
	application := app.NewApp()

	if err := application.InstallDependencies(app.PostgresDependency); err != nil {
		log.Fatalf("Failed to install dependencies: %v", err)
	}

	defer application.Close()

	registerUserService := func(srv *grpc.Server, app *app.App) {
		userRepo := repository.NewUserRepository(app.Store.Postgres.DB)

		userUsecase := usecase.NewUserUsecase(userRepo)

		server.RegisterService(srv, userUsecase)
	}

	interceptorOpt := grpc.UnaryInterceptor(interceptors.LoggingInterceptor)

	application.RunGRPCServer(constants.UserServiceName, registerUserService, interceptorOpt)
}
