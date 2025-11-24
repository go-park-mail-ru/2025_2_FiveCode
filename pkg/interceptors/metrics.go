package interceptors

import (
	"context"
	"time"

	"backend/pkg/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func MetricsInterceptor(serviceName string) grpc.UnaryServerInterceptor {
	m := metrics.NewGRPCMetrics(serviceName)

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		method := info.FullMethod

		m.IncreaseHits(method)

		resp, err := handler(ctx, req)

		duration := time.Since(start)
		m.RecordResponseTime(method, duration.Seconds())

		if err != nil {
			st, _ := status.FromError(err)
			code := int(st.Code())
			m.IncreaseErr(method, code)
		}

		return resp, err
	}
}
