package exporter

import (
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

var name = "Server"

func newExporter(url string) (trace.SpanExporter, error) {
	return jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
}

func newResource(name string) *resource.Resource {
	r, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(name),
			semconv.ServiceVersionKey.String("v0.1.0"),
			attribute.String("environment", "demo"),
		),
	)
	return r
}
func newTraceProvider(exporter trace.SpanExporter) (*trace.TracerProvider, error) {
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(newResource(name)),
	)

	return tp, nil
}

func Register() {
	exp, err := newExporter("http://localhost:14268/api/traces")

	if err != nil {
		log.Fatalf("Register exporter: %v", err)
	}

	tp, err := newTraceProvider(exp)

	if err != nil {
		log.Fatal(err)
	}

	otel.SetTracerProvider(tp)
}
