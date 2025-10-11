package app

import (
	"backend/config"
	"backend/router"
	"backend/store"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"net/http"
)

func RunApp() error {
	s := store.NewStore()

	if err := s.InitFillStore(); err != nil {
		return fmt.Errorf("failed to init store: %w", err)
	}

	r := router.NewRouter(s)

	serverAddr, err := config.ReadServerAddress()
	if err != nil {
		return fmt.Errorf("failed to read server address: %w", err)
	}

	server := &http.Server{
		Addr:    serverAddr,
		Handler: r,
	}

	log.Info().Str("addr", server.Addr).Msg("listening")
	err = server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}
