// TODO: make crsf protection for notes service

// func (u *AuthUsecase) GenerateCSRFToken(ctx context.Context, sessionID string) (string, error) {
// 	log := logger.FromContext(ctx)
// 	log.Info().Str("session_id", sessionID).Msg("generating csrf token")

// 	token, err := apiutils.GenerateCSRFToken(sessionID, u.CSRFSecret)
// 	if err != nil {
// 		log.Error().Err(err).Msg("failed to generate csrf token")
// 		return "", fmt.Errorf("failed to generate csrf token: %w", err)
// 	}

// 	return token, nil
// }

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