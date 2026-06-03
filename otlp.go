package main

// otlp.go adds an OTLP metrics push path alongside the existing Prometheus
// /metrics endpoint. It reuses this exporter's existing Prometheus registry via
// the OpenTelemetry Prometheus bridge and periodically ships those metrics to an
// OTLP collector (OTel Collector -> Telegraf -> TimescaleDB). The /metrics
// endpoint is unaffected, so the exporter supports both backends at once.

import (
	"context"
	"os"
	"strings"
	"time"

	promclient "github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	otelprom "go.opentelemetry.io/contrib/bridges/prometheus"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// otlpServiceName is recorded as the service.name resource attribute, which
// becomes a tag column in TimescaleDB. Override at runtime with
// OTLP_SERVICE_NAME.
const otlpServiceName = "nws_exporter"

// otlpDropPrefixes are metric-name prefixes for runtime/exporter-internal
// metrics that should not be shipped to TimescaleDB.
var otlpDropPrefixes = []string{"go_", "process_", "promhttp_"}

// filteringGatherer wraps a Prometheus gatherer and drops metric families whose
// names match otlpDropPrefixes, keeping runtime noise out of TimescaleDB.
type filteringGatherer struct{ g promclient.Gatherer }

func (f filteringGatherer) Gather() ([]*dto.MetricFamily, error) {
	mfs, err := f.g.Gather()
	if err != nil {
		return nil, err
	}
	out := mfs[:0]
	for _, mf := range mfs {
		drop := false
		for _, p := range otlpDropPrefixes {
			if strings.HasPrefix(mf.GetName(), p) {
				drop = true
				break
			}
		}
		if !drop {
			out = append(out, mf)
		}
	}
	return out, nil
}

// startOTLPPush starts periodically pushing the metrics from the given gatherer
// to an OTLP endpoint. Configuration via env vars:
//
//	OTLP_ENDPOINT      host:port of the OTLP/HTTP collector (default 127.0.0.1:4318)
//	OTLP_SERVICE_NAME  service.name resource attribute (default otlpServiceName)
//	OTLP_INTERVAL      push interval as a Go duration (default 60s)
//
// It returns a stop function that flushes and shuts down the exporter.
func startOTLPPush(gatherer promclient.Gatherer) (func(), error) {
	ctx := context.Background()

	endpoint := os.Getenv("OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "127.0.0.1:4318"
	}
	service := os.Getenv("OTLP_SERVICE_NAME")
	if service == "" {
		service = otlpServiceName
	}
	interval := 60 * time.Second
	if v := os.Getenv("OTLP_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			interval = d
		}
	}

	exp, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpoint(endpoint),
		otlpmetrichttp.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(semconv.ServiceName(service)),
	)
	if err != nil {
		return nil, err
	}

	reader := metric.NewPeriodicReader(exp,
		metric.WithInterval(interval),
		metric.WithProducer(otelprom.NewMetricProducer(
			otelprom.WithGatherer(filteringGatherer{gatherer}),
		)),
	)
	mp := metric.NewMeterProvider(
		metric.WithReader(reader),
		metric.WithResource(res),
	)
	return func() { _ = mp.Shutdown(context.Background()) }, nil
}
