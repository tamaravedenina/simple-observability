package internal

import (
	muxtrace "go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux"
	"go.uber.org/zap"
	"net"
	"net/http"

	"github.com/gorilla/mux"
	oteltrace "go.opentelemetry.io/otel/api/trace"
)

// Diagnostics responsible for diagnostics logic of the app
func Diagnostics(port string, appLogger *zap.SugaredLogger, tracer oteltrace.Tracer, metricsHandler func(http.ResponseWriter, *http.Request), shutdown chan<- error) *http.Server {
	r := mux.NewRouter()

	mw := muxtrace.Middleware("diag", muxtrace.WithTracer(tracer))
	r.Use(mw)

	r.HandleFunc("/health", handleHealth(appLogger.With("handle", "health")))
	r.HandleFunc("/metrics", metricsHandler)

	server := http.Server{
		Addr:    net.JoinHostPort("", port),
		Handler: r,
	}

	appLogger.Info("Ready to start the server...")

	go func() {
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			shutdown <- err
		}
	}()

	return &server
}

func handleHealth(appLogger *zap.SugaredLogger) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		appLogger.Info("Received a call")
		w.WriteHeader(http.StatusOK)
	}
}
