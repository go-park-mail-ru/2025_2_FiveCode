package metrics

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type metrics struct {
	httpHits          *prometheus.CounterVec
	httpErrors        *prometheus.CounterVec
	httpResponseTime  *prometheus.HistogramVec
	grpcHits          *prometheus.CounterVec
	grpcErrors        *prometheus.CounterVec
	grpcResponseTime  *prometheus.HistogramVec
}

var (
	defaultMetrics *metrics
)

func init() {
	defaultMetrics = &metrics{
		httpHits: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "service"},
		),
		httpErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_errors_total",
				Help: "Total number of HTTP errors",
			},
			[]string{"method", "path", "code", "service"},
		),
		httpResponseTime: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path", "service"},
		),
		grpcHits: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "grpc_requests_total",
				Help: "Total number of gRPC requests",
			},
			[]string{"method", "service"},
		),
		grpcErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "grpc_errors_total",
				Help: "Total number of gRPC errors",
			},
			[]string{"method", "code", "service"},
		),
		grpcResponseTime: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "grpc_request_duration_seconds",
				Help:    "gRPC request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "service"},
		),
	}
}

type httpMetrics struct {
	service string
	m       *metrics
}

func NewHTTPMetrics(serviceName string) *httpMetrics {
	return &httpMetrics{
		service: serviceName,
		m:       defaultMetrics,
	}
}

func (h *httpMetrics) IncreaseHits(method, path string) {
	h.m.httpHits.WithLabelValues(method, path, h.service).Inc()
}

func (h *httpMetrics) IncreaseErr(method, path string, code int) {
	codeStr := strconv.Itoa(code)
	h.m.httpErrors.WithLabelValues(method, path, codeStr, h.service).Inc()
}

func (h *httpMetrics) RecordResponseTime(method, path string, durationSeconds float64) {
	h.m.httpResponseTime.WithLabelValues(method, path, h.service).Observe(durationSeconds)
}

type grpcMetrics struct {
	service string
	m       *metrics
}

func NewGRPCMetrics(serviceName string) *grpcMetrics {
	return &grpcMetrics{
		service: serviceName,
		m:       defaultMetrics,
	}
}

func (g *grpcMetrics) IncreaseHits(method string) {
	g.m.grpcHits.WithLabelValues(method, g.service).Inc()
}

func (g *grpcMetrics) IncreaseErr(method string, code int) {
	codeStr := strconv.Itoa(code)
	g.m.grpcErrors.WithLabelValues(method, codeStr, g.service).Inc()
}

func (g *grpcMetrics) RecordResponseTime(method string, durationSeconds float64) {
	g.m.grpcResponseTime.WithLabelValues(method, g.service).Observe(durationSeconds)
}
