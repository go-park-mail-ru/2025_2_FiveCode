package logger

import (
	"context"
	"regexp"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type ctxKey string

const loggerKey ctxKey = "logger"

var whiteSpaceRegex = regexp.MustCompile(`\s+`)

func SanitizeQuery(q string) string {
	singleSpaceQuery := whiteSpaceRegex.ReplaceAllString(q, " ")
	return strings.TrimSpace(singleSpaceQuery)
}

func ToContext(ctx context.Context, logger zerolog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

func FromContext(ctx context.Context) zerolog.Logger {
	logger, ok := ctx.Value(loggerKey).(zerolog.Logger)
	if !ok {
		return log.Logger
	}
	return logger
}
