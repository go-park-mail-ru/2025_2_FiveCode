package main

import (
	"backend/internal/app"
	"backend/internal/initialize"
	"backend/internal/router"
	"github.com/rs/zerolog/log"
	"net/http"
)

func main() {
	application := app.NewApp()

	defer func() {
		if err := application.Close(); err != nil {
			log.Printf("Error closing app resources: %v", err)
		}
	}()

	if err := application.InstallDependencies(
		app.PostgresDependency,
		app.MinioDependency,
	); err != nil {
		log.Fatal().Err(err).Msg("Failed to install dependencies")
		return
	}

	deliveries := initialize.InitDeliveries(application.Store, application.Config)

	mainRouter := router.NewRouter(application.Store, deliveries, application.Config)

	application.RunHTTPServer(func(app *app.App) http.Handler {
		return mainRouter
	})
}
