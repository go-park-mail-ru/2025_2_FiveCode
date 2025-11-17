package main

import (
	"backend/app"
	"backend/initialize"
	"backend/router"
	"net/http"

	"github.com/rs/zerolog/log"
)

// @title Goose API
// @version 1.0
// @description API for web-site Goose
// @BasePath /api
// @securityDefinitions.apikey CookieAuth
// @in header
// @name Cookie
func main() {
	application := app.NewApp()

	defer func() {
		if err := application.Close(); err != nil {
			log.Printf("Error closing app resources: %v", err)
		}
	}()

	if err := application.InstallDependencies(
		app.PostgresDependency,
		app.RedisDependency,
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
