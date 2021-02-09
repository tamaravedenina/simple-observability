package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tamaravedenina/observability/internal"
	otelg "go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/exporters/stdout"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
)

const serviceName = "simple-observability"

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	appLogger := logger.Sugar().Named(serviceName)
	appLogger.Info("The application is starting...")

	exporter, err := stdout.NewExporter(stdout.WithPrettyPrint())
	if err != nil {
		appLogger.Fatalw("Can't enable Open Telemetry exporter", "err", err)
	}

	tp, err := sdktrace.NewProvider(
		sdktrace.WithConfig(
			sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()},
		),
		sdktrace.WithSyncer(exporter),
	)

	if err != nil {
		appLogger.Fatalw("Can't enable Open Telemetry provider", "err", err)
	}
	otelg.SetTraceProvider(tp)

	tracer := otelg.Tracer(serviceName)

	appLogger.Info("Reading configuration...")
	port := os.Getenv("PORT")
	if port == "" {
		appLogger.Fatal("PORT is not set...")
	}

	diagPort := os.Getenv("DIAG_PORT")
	if diagPort == "" {
		appLogger.Fatal("DIAG_PORT is not set...")
	}

	appLogger.Info("Configuration is ready...")

	shutdown := make(chan error, 2)
	bl := internal.BusinessLogic(port, appLogger.With("module", "bl"), tracer, shutdown)
	diag := internal.Diagnostics(diagPort, appLogger.With("module", "diag"), tracer, shutdown)
	appLogger.Info("Servers are ready")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case x := <-interrupt:
		appLogger.Infow("Received", "signal", x.String())
	case err := <-shutdown:
		appLogger.Errorw("Received error from functional unit", "err", err)
	}

	appLogger.Info("Stopping the servers...")
	timeout, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()

	err = bl.Shutdown(timeout)
	if err != nil {
		appLogger.Errorw("Got an error from the business logic server", "err", err)
	}
	err = diag.Shutdown(timeout)
	if err != nil {
		appLogger.Errorw("Got an error from the business diagnostic server", "err", err)
	}

	appLogger.Info("The application is stopped")
}
