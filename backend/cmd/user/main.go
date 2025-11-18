package main

import (
	"backend/app"
	"backend/pkg/interceptors"
	"backend/pkg/user/repository"
	"backend/pkg/user/server"
	"backend/pkg/user/usecase"
	"log"

	"google.golang.org/grpc"
)

const (
	userServiceName = "users"
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

	application.RunGRPCServer(userServiceName, registerUserService, interceptorOpt)
}
