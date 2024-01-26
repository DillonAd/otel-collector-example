package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())

	otel := NewOtel(ctx)
	defer otel.Close(ctx)

	go generateTelemetry(ctx)
	<-shutdown
	fmt.Printf("shutdown")
	cancel()
}
