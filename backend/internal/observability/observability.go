package observability

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/udaykishore-resu/travelmind/internal/config"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"net/http"
)

// NewLogger creates a new logrus logger
func NewLogger(logLevel string) *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	return logger
}

// InitTracer initializes OpenTelemetry tracing
func InitTracer(cfg *config.Config) (*sdktrace.TracerProvider, error) {
	exporter, err := otlptracehttp.New(context.Background(),
		otlptracehttp.WithEndpoint("localhost:4318"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	resources, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName("travelmind-api-gateway"),
			semconv.ServiceVersion(cfg.Version),
			semconv.DeploymentEnvironment(cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resources),
	)

	otel.SetTracerProvider(tp)

	return tp, nil
}

// Metrics holds Prometheus metrics
type Metrics struct {
	httpRequestsTotal   prometheus.Counter
	httpRequestDuration prometheus.Histogram
	dbQueryDuration     prometheus.Histogram
	dbConnections       prometheus.Gauge
}

// InitMetrics initializes Prometheus metrics
func InitMetrics() http.Handler {
	// Register metrics
	httpRequestsTotal := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "travelmind_http_requests_total",
		Help: "Total number of HTTP requests",
	})

	httpRequestDuration := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "travelmind_http_request_duration_seconds",
		Help:    "HTTP request duration in seconds",
		Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
	})

	dbQueryDuration := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "travelmind_db_query_duration_seconds",
		Help:    "Database query duration in seconds",
		Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
	})

	dbConnections := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "travelmind_db_connections",
		Help: "Number of active database connections",
	})

	// Register metrics
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(dbQueryDuration)
	prometheus.MustRegister(dbConnections)

	return promhttp.Handler()
}

// HealthCheckResult represents health check result
type HealthCheckResult struct {
	Status    string                 `json:"status"`
	Checks    map[string]CheckResult `json:"checks"`
	Uptime    int64                  `json:"uptime"`
	Timestamp string                 `json:"timestamp"`
}

// CheckResult represents a single health check result
type CheckResult struct {
	Status  string      `json:"status"`
	Details interface{} `json:"details,omitempty"`
	Error   string      `json:"error,omitempty"`
}
