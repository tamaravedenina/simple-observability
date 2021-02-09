package internal

import (
	muxtrace "go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux"
	"go.opentelemetry.io/otel/instrumentation/httptrace"
	oteltrace "go.opentelemetry.io/otel/api/trace"
	"go.uber.org/zap"
	"net"
	"net/http"

	"github.com/gorilla/mux"
)

func BusinessLogic(port string, appLogger *zap.SugaredLogger, tracer oteltrace.Tracer, shutdown chan<- error) *http.Server {
	r := mux.NewRouter()

	mw := muxtrace.Middleware("bl", muxtrace.WithTracer(tracer))
	r.Use(mw)

	r.HandleFunc("/rent", handleRent(appLogger.With("handle", "rent"), "http://127.0.0.1:"+port+"/check"))
	r.HandleFunc("/check", handleCheck(appLogger.With("handle", "check")))

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

func handleRent(appLogger *zap.SugaredLogger, checkURL string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		appLogger.Info("Received a call")

		req, err := http.NewRequest(http.MethodGet, checkURL, nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			appLogger.Errorw("Error when creating request to check", "err", err)
			return
		}

		httptrace.Inject(r.Context(), req)
		checkResp, err := http.DefaultClient.Do(req)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			appLogger.Errorw("Error when do request to check", "err", err)
			return
		}

		w.WriteHeader(checkResp.StatusCode)
	}
}

func handleCheck(appLogger *zap.SugaredLogger) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		appLogger.Info("Received a call")
		// extract
		w.WriteHeader(http.StatusOK)
	}
}
