package metrics

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestNewHTTPMetrics(t *testing.T) {
	m := NewHTTPMetrics("test_service")
	assert.NotNil(t, m)
	assert.Equal(t, "test_service", m.service)
	assert.NotNil(t, m.m)
}

func TestHTTPMetrics_IncreaseHits(t *testing.T) {
	m := NewHTTPMetrics("test_service")
	m.IncreaseHits("GET", "/")
}

func TestHTTPMetrics_IncreaseErr(t *testing.T) {
	m := NewHTTPMetrics("test_service")
	m.IncreaseErr("GET", "/", 500)
}

func TestHTTPMetrics_RecordResponseTime(t *testing.T) {
	m := NewHTTPMetrics("test_service")
	m.RecordResponseTime("GET", "/", 0.1)
}

func TestNewGRPCMetrics(t *testing.T) {
	m := NewGRPCMetrics("test_service")
	assert.NotNil(t, m)
	assert.Equal(t, "test_service", m.service)
	assert.NotNil(t, m.m)
}

func TestGRPCMetrics_IncreaseHits(t *testing.T) {
	m := NewGRPCMetrics("test_service")
	m.IncreaseHits("GetProfile")
}

func TestGRPCMetrics_IncreaseErr(t *testing.T) {
	m := NewGRPCMetrics("test_service")
	m.IncreaseErr("GetProfile", 13)
}

func TestGRPCMetrics_RecordResponseTime(t *testing.T) {
	m := NewGRPCMetrics("test_service")
	m.RecordResponseTime("GetProfile", 0.05)
}

func TestRegistry(t *testing.T) {
	r := Registry()
	assert.NotNil(t, r)
	assert.IsType(t, &prometheus.Registry{}, r)
}

func TestRecordDBQueryDuration(t *testing.T) {
	start := time.Now()
	RecordDBQueryDuration(start, "select", "users")
}

func TestRecordDBQueryError(t *testing.T) {
	RecordDBQueryError("insert", "users")
}
