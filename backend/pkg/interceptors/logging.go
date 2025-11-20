package interceptors

import (
	"backend/pkg/logger"
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func LoggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	log := logger.FromContext(ctx)
	start := time.Now()

	log.Info().Str("grpc_method", info.FullMethod).Msg("gRPC request started")

	resp, err := handler(ctx, req)

	duration := time.Since(start)

	if err != nil {
		st, _ := status.FromError(err)
		log.Error().Err(err).Str("grpc_method", info.FullMethod).Str("grpc_code", st.Code().String()).Dur("duration", duration).Msg("gRPC request failed")
	} else {
		log.Info().Str("grpc_method", info.FullMethod).Dur("duration", duration).Msg("gRPC request finished successfully")
	}

	return resp, err
}
