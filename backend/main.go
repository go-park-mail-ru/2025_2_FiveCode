package main

import (
	"backend/app"
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
	err := app.RunApp()
	if err != nil {
		log.Fatal().Err(err).Msg("Error running app")
	}
}
