package main

import (
	"backend/app"
	"backend/constants"
	"backend/pkg/interceptors"
	"backend/pkg/user/repository"
	"backend/pkg/user/server"
	"backend/pkg/user/usecase"
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
