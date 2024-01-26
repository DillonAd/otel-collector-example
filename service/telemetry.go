package main

import (
	"context"
	"math/rand"
	"time"

	"github.com/robfig/cron/v3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	oteltrace "go.opentelemetry.io/otel/trace"
)

var tracer oteltrace.Tracer
var counter metric.Int64Counter
var duration metric.Float64Histogram

func generateTelemetry(ctx context.Context) {
	tracer = otel.GetTracerProvider().Tracer("custom-service")
	var meter = otel.Meter("custom-service")

	var err error
	counter, err = meter.Int64Counter("telemetry-run")

	if err != nil {
		panic(err)
	}

	duration, err = meter.Float64Histogram(
		"telemetry-run-duration",
		metric.WithDescription("The duration of task execution."),
		metric.WithUnit("s"),
	)

	if err != nil {
		panic(err)
	}

	c := cron.New()

	for i := 0; i < 100; i++ {
		_, err = c.AddFunc("* * * * *", run)
		if err != nil {
			panic(err)
		}
	}

	c.Start()

	<-ctx.Done()
	c.Stop()
}

func run() {
	start := time.Now().UTC()
	spanCtx, span := tracer.Start(context.Background(), "run")
	defer span.End()

	counter.Add(spanCtx, 1)

	interval := rand.Intn(60000)
	span.SetAttributes(
		attribute.Int("interval", interval),
	)

	run_subtask(spanCtx)

	time.Sleep(time.Millisecond * time.Duration(interval))
	duration.Record(spanCtx, time.Since(start).Seconds())
}

func run_subtask(ctx context.Context) {
	_, span := tracer.Start(ctx, "run_subtask")
	defer span.End()

	interval := rand.Intn(10000)
	span.SetAttributes(
		attribute.Int("interval", interval),
	)
	time.Sleep(time.Millisecond * time.Duration(interval))
}