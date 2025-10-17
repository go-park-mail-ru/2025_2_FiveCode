package app

import (
	"backend/config"
	"backend/initialize"
	"backend/router"
	"backend/store"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
		NoColor:    false,
	})
}

func RunApp() error {
	s := store.NewStore()

	if err := s.InitFillStore(); err != nil {
		return fmt.Errorf("failed to fill store: %w", err)
	}

	configPath, err := config.ReadConfigPath()
	if err != nil {
		return fmt.Errorf("failed to read config path: %w", err)
	}

	conf, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	deliveries := initialize.InitDeliveries(s, conf)

	r := router.NewRouter(s, deliveries)

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
