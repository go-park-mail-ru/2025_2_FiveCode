package main

import (
	"backend/notes_service/app"

	"github.com/rs/zerolog/log"
)

func main() {
	application := app.NewApp()

	defer func() {
		if err := application.Close(); err != nil {
			log.Error().Err(err).Msg("Error closing app resources")
		}
	}()

	application.Run()
}
