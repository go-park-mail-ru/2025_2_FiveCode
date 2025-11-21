package interceptors

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const RequestIDKey = "request-id"

type ToContextFunc func(ctx context.Context, logger zerolog.Logger) context.Context
type FromContextFunc func(ctx context.Context) zerolog.Logger

func LoggingInterceptor(baseLogger zerolog.Logger, toCtx ToContextFunc) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		var requestID string
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if val, ok := md[RequestIDKey]; ok && len(val) > 0 {
				requestID = val[0]
			}
		}

		if requestID == "" {
			requestID = uuid.New().String()
		}

		reqLogger := baseLogger.With().Str("request_id", requestID).Logger()
		ctx = toCtx(ctx, reqLogger)

		reqLogger.Info().Str("grpc_method", info.FullMethod).Msg("gRPC request started")

		resp, err := handler(ctx, req)

		duration := time.Since(start)

		if err != nil {
			st, _ := status.FromError(err)
			reqLogger.Error().Err(err).Str("grpc_method", info.FullMethod).Str("grpc_code", st.Code().String()).Dur("duration", duration).Msg("gRPC request failed")
		} else {
			reqLogger.Info().Str("grpc_method", info.FullMethod).Dur("duration", duration).Msg("gRPC request finished successfully")
		}

		return resp, err
	}
}