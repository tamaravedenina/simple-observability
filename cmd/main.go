package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/tamaravedenina/observability/internal"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	appLogger := logger.Sugar().Named("observability")
	appLogger.Info("The application is starting...")

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
	bl := internal.BusinessLogic(port, appLogger.With("module", "bl"), shutdown)
	diag := internal.Diagnostics(diagPort, appLogger.With("module", "diag"), shutdown)
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

	err := bl.Shutdown(timeout)
	if err != nil {
		appLogger.Errorw("Got an error from the business logic server", "err", err)
	}
	err = diag.Shutdown(timeout)
	if err != nil {
		appLogger.Errorw("Got an error from the business diagnostic server", "err", err)
	}

	appLogger.Info("The application is stopped")
}
