package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type Otel struct {
	tracerProvider *sdktrace.TracerProvider
	meterProvider  *metric.MeterProvider
}

func NewOtel(ctx context.Context) *Otel {
	endpoint := os.Getenv("OTEL_COLLECTOR_ENDPOINT")

	if endpoint == "" {
		log.Println("no opentelemetry collector endpoint")
		return nil
	}

	// Start Tracing

	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	}

	traceExporter, err := otlptracegrpc.New(ctx, opts...)
	if err != nil {
		panic(fmt.Errorf("error creating trace exporter: %v", err))
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("custom_service"),
		),
		resource.WithFromEnv(),
	)

	if err != nil {
		panic(fmt.Errorf("error creating tracing resource: %v", err))
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	otel.SetTracerProvider(tracerProvider)

	// End Tracing

	// Start Metics

	metricExporter, err := otlpmetricgrpc.New(
		ctx,
		otlpmetricgrpc.WithEndpoint(endpoint),
		otlpmetricgrpc.WithInsecure(),
	)

	if err != nil {
		panic(fmt.Errorf("error creating metrics exporter: %v", err))
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(metricExporter,
			metric.WithInterval(3*time.Second))),
	)

	otel.SetMeterProvider(meterProvider)

	// End Metrics

	return &Otel{
		tracerProvider: tracerProvider,
		meterProvider:  meterProvider,
	}
}

func (o *Otel) GetTracer() oteltrace.Tracer {
	return otel.GetTracerProvider().Tracer("custom-service")
}

func (o *Otel) Close(ctx context.Context) {
	if err := o.tracerProvider.Shutdown(ctx); err != nil {
		log.Printf("error stopping tracer provider: %v\n", err)
	}
	if err := o.meterProvider.Shutdown(ctx); err != nil {
		log.Printf("error stopping metric provider: %v\n", err)
	}
}
