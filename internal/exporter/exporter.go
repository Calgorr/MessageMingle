package exporter

import (
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

// Config holds configuration parameters for the exporter.
type Config struct {
	CollectorEndpoint string
	ServiceName       string
	ServiceVersion    string
	Environment       string
}

func newExporter(config *Config) (trace.SpanExporter, error) {
	endpoint := fmt.Sprintf("http://%s/api/traces", config.CollectorEndpoint)
	return jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(endpoint)))
}

func newResource(config *Config) *resource.Resource {
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(config.ServiceName),
		semconv.ServiceVersionKey.String(config.ServiceVersion),
		attribute.String("environment", config.Environment),
	)
}

func newTraceProvider(exporter trace.SpanExporter, config *Config) (*trace.TracerProvider, error) {
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(newResource(config)),
	)

	return tp, nil
}

// Register registers the exporter as the global trace provider.
// If config is nil, default values will be used.
func Register(config *Config) error {
	if config == nil {
		config = &Config{
			CollectorEndpoint: "localhost:14268",
			ServiceName:       "therealbroker",
			ServiceVersion:    "0.0.1",
			Environment:       "development",
		}
	}
	exp, err := newExporter(config)
	if err != nil {
		return err
	}

	tp, err := newTraceProvider(exp, config)
	if err != nil {
		return err
	}

	otel.SetTracerProvider(tp)
	return nil
}
