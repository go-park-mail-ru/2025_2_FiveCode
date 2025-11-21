package logger

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

type ctxKey string

const loggerKey ctxKey = "logger"

func New() zerolog.Logger {
	logLevelStr := viper.GetString("LOG_LEVEL")
	if logLevelStr == "" {
		logLevelStr = "info"
	}

	logLevel, err := zerolog.ParseLevel(strings.ToLower(logLevelStr))
	if err != nil {
		logLevel = zerolog.InfoLevel
	}

	return zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	}).Level(logLevel).With().Timestamp().Str("service", "notes-service").Logger()
}

func ToContext(ctx context.Context, logger zerolog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

func FromContext(ctx context.Context) zerolog.Logger {
	if logger, ok := ctx.Value(loggerKey).(zerolog.Logger); ok {
		return logger
	}
	return zerolog.Nop()
}