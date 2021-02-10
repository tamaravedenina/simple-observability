package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	otelg "go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/exporters/metric/prometheus"
	"go.opentelemetry.io/otel/exporters/stdout"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/sdk/metric/controller/push"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"

	"github.com/tamaravedenina/observability/internal"
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

	jaegerEndpoint := os.Getenv("JAEGER_ENDPOINT")
	if jaegerEndpoint == "" {
		appLogger.Fatal("Jaeger endpoint is not set")
	}
	jExporter, err := jaeger.NewRawExporter(
		jaeger.WithAgentEndpoint(jaegerEndpoint),
		jaeger.WithProcess(jaeger.Process{ServiceName: serviceName}),
	)
	if err != nil {
		appLogger.Fatalw("Can't set Jaeger exporter", "err", err)
	}

	tp, err := sdktrace.NewProvider(
		sdktrace.WithConfig(
			sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()},
		),
		sdktrace.WithSyncer(exporter),
		sdktrace.WithSyncer(jExporter),
	)
	if err != nil {
		appLogger.Fatalw("Can't set Open Telemetry provider", "err", err)
	}
	otelg.SetTraceProvider(tp)

	tracer := otelg.Tracer(serviceName)

	mc := push.New(
		simple.NewWithExactDistribution(),
		exporter,
	)
	mc.Start()
	defer mc.Stop()
	otelg.SetMeterProvider(mc.Provider())

	pExporter, err := prometheus.NewExportPipeline(prometheus.Config{})
	if err != nil {
		appLogger.Fatalw("Can't set Prometheus exporter", "err", err)
	}
	otelg.SetMeterProvider(pExporter.Provider())

	meter := otelg.Meter(serviceName)

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
	bl := internal.BusinessLogic(port, appLogger.With("module", "bl"), tracer, meter, shutdown)
	diag := internal.Diagnostics(diagPort, appLogger.With("module", "diag"), tracer, pExporter.ServeHTTP, shutdown)
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
