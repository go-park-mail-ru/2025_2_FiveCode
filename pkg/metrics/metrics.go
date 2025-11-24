package metrics

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type metrics struct {
	httpHits         *prometheus.CounterVec
	httpErrors       *prometheus.CounterVec
	httpResponseTime *prometheus.HistogramVec
	grpcHits         *prometheus.CounterVec
	grpcErrors       *prometheus.CounterVec
	grpcResponseTime *prometheus.HistogramVec
	dbQueryDuration  *prometheus.HistogramVec
	dbQueryErrors    *prometheus.CounterVec
}

var (
	defaultMetrics *metrics
	registry       = prometheus.NewRegistry()
)

func init() {
	registry.MustRegister(
		prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
		prometheus.NewGoCollector(),
	)

	factory := promauto.With(registry)

	defaultMetrics = &metrics{
		httpHits: factory.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "service"},
		),
		httpErrors: factory.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_errors_total",
				Help: "Total number of HTTP errors",
			},
			[]string{"method", "path", "code", "service"},
		),
		httpResponseTime: factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path", "service"},
		),
		grpcHits: factory.NewCounterVec(
			prometheus.CounterOpts{
				Name: "grpc_requests_total",
				Help: "Total number of gRPC requests",
			},
			[]string{"method", "service"},
		),
		grpcErrors: factory.NewCounterVec(
			prometheus.CounterOpts{
				Name: "grpc_errors_total",
				Help: "Total number of gRPC errors",
			},
			[]string{"method", "code", "service"},
		),
		grpcResponseTime: factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "grpc_request_duration_seconds",
				Help:    "gRPC request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "service"},
		),
		dbQueryDuration: factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "db_query_duration_seconds",
				Help:    "Database query duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"operation", "table"},
		),
		dbQueryErrors: factory.NewCounterVec(
			prometheus.CounterOpts{
				Name: "db_query_errors_total",
				Help: "Total number of database query errors",
			},
			[]string{"operation", "table"},
		),
	}
}

func Registry() *prometheus.Registry {
	return registry
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

func RecordDBQueryDuration(start time.Time, operation, table string) {
	duration := time.Since(start).Seconds()
	defaultMetrics.dbQueryDuration.WithLabelValues(operation, table).Observe(duration)
}

func RecordDBQueryError(operation, table string) {
	defaultMetrics.dbQueryErrors.WithLabelValues(operation, table).Inc()
}
