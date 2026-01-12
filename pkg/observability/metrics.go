package observability

import (
	"context"
	"fmt"

	gcpmetric "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric"
	"github.com/gofiber/fiber/v2/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

const (
	meterName = "driftive.cloud/api"
)

// Metrics holds all application metrics
type Metrics struct {
	meter metric.Meter

	// Token refresh metrics
	TokenRefreshTotal     metric.Int64Counter
	TokenRefreshSuccess   metric.Int64Counter
	TokenRefreshFailure   metric.Int64Counter
	TokenRefreshDisabled  metric.Int64Counter
	TokenRefreshRateLimit metric.Int64Counter
}

// metricsInstance is the singleton instance
var metricsInstance *Metrics

// GetMetrics returns the metrics singleton
func GetMetrics() *Metrics {
	return metricsInstance
}

// InitMetrics initializes the OpenTelemetry metrics with GCP exporter
func InitMetrics(ctx context.Context, projectID string) (*Metrics, func(), error) {
	// Create GCP exporter
	exporter, err := gcpmetric.New(gcpmetric.WithProjectID(projectID))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create GCP metric exporter: %w", err)
	}

	// Create meter provider with the exporter
	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter)),
	)
	otel.SetMeterProvider(provider)

	// Create meter
	meter := provider.Meter(meterName)

	// Initialize metrics
	metrics, err := createMetrics(meter)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create metrics: %w", err)
	}

	metricsInstance = metrics

	// Return shutdown function
	shutdown := func() {
		if err := provider.Shutdown(ctx); err != nil {
			log.Errorf("error shutting down meter provider: %v", err)
		}
	}

	return metrics, shutdown, nil
}

// InitMetricsNoop initializes a no-op metrics instance for local development
func InitMetricsNoop() *Metrics {
	meter := otel.Meter(meterName)
	metrics, err := createMetrics(meter)
	if err != nil {
		log.Warnf("failed to create noop metrics: %v", err)
		return nil
	}
	metricsInstance = metrics
	return metrics
}

func createMetrics(meter metric.Meter) (*Metrics, error) {
	tokenRefreshTotal, err := meter.Int64Counter(
		"token_refresh_total",
		metric.WithDescription("Total number of token refresh attempts"),
	)
	if err != nil {
		return nil, err
	}

	tokenRefreshSuccess, err := meter.Int64Counter(
		"token_refresh_success",
		metric.WithDescription("Number of successful token refreshes"),
	)
	if err != nil {
		return nil, err
	}

	tokenRefreshFailure, err := meter.Int64Counter(
		"token_refresh_failure",
		metric.WithDescription("Number of failed token refreshes"),
	)
	if err != nil {
		return nil, err
	}

	tokenRefreshDisabled, err := meter.Int64Counter(
		"token_refresh_disabled",
		metric.WithDescription("Number of users disabled due to repeated failures"),
	)
	if err != nil {
		return nil, err
	}

	tokenRefreshRateLimit, err := meter.Int64Counter(
		"token_refresh_rate_limited",
		metric.WithDescription("Number of rate limit responses from GitHub"),
	)
	if err != nil {
		return nil, err
	}

	return &Metrics{
		meter:                 meter,
		TokenRefreshTotal:     tokenRefreshTotal,
		TokenRefreshSuccess:   tokenRefreshSuccess,
		TokenRefreshFailure:   tokenRefreshFailure,
		TokenRefreshDisabled:  tokenRefreshDisabled,
		TokenRefreshRateLimit: tokenRefreshRateLimit,
	}, nil
}
