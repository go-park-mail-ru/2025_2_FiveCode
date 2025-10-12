package app

import (
	"backend/config"
	"backend/router"
	"backend/store"
	"fmt"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"time"
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
		return errors.Wrap(err, "failed to initially fill store")
	}

	configPath, err := config.ReadConfigPath()
	if err != nil {
		return errors.Wrap(err, "failed to read config path")
	}

	conf, err := config.LoadConfig(configPath)
	if err != nil {
		return errors.Wrap(err, "failed to load config file")
	}

	r := router.NewRouter(s, conf)

	serverAddr, err := config.ReadServerAddress()
	if err != nil {
		return errors.Wrap(err, "failed to read server address")
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
