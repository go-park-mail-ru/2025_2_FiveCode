package main

import (
	"backend/gateway_service/app"
	"backend/gateway_service/initialize"
	"backend/gateway_service/router"
	"net/http"

	"github.com/rs/zerolog/log"
)

func main() {
	application := app.NewApp()

	defer func() {
		if err := application.Close(); err != nil {
			log.Error().Err(err).Msgf("Error closing app resources: %v", err)
		}
	}()

	deliveries := initialize.InitDeliveries(application.Store, application.Config)

	mainRouter := router.NewRouter(application.Store, deliveries, application.Config, &application.Logger)

	application.RunHTTPServer(func(app *app.App) http.Handler {
		return mainRouter
	})
}
