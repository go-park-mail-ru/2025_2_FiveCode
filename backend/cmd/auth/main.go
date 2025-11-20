package main

import (
	"backend/app"
	"backend/constants"
	"backend/pkg/auth/repository"
	"backend/pkg/auth/server"
	"backend/pkg/auth/usecase"
	"backend/pkg/interceptors"
	"log"

	"google.golang.org/grpc"
)

func main() {
	application := app.NewApp()

	if err := application.InstallDependencies(app.PostgresDependency, app.RedisDependency); err != nil {
		log.Fatalf("Failed to install dependencies: %v", err)
	}

	defer application.Close()

	registerAuthService := func(srv *grpc.Server, app *app.App) {
		repo := repository.NewAuthRepository(app.Store.Postgres.DB, app.Store.Redis.Client)
		useCase := usecase.NewAuthUsecase(repo, []byte(app.Config.Auth.CSRF.SecretKey))

		server.RegisterService(srv, useCase)
	}

	interceptorOpt := grpc.UnaryInterceptor(interceptors.LoggingInterceptor)

	application.RunGRPCServer(constants.AuthServiceName, registerAuthService, interceptorOpt)
}
