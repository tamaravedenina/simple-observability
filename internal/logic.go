package internal

import (
	"go.uber.org/zap"
	"net"
	"net/http"

	"github.com/gorilla/mux"
)

func BusinessLogic(port string, appLogger *zap.SugaredLogger, shutdown chan<- error) *http.Server {
	r := mux.NewRouter()
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
	return func(
		w http.ResponseWriter, r *http.Request) {
		appLogger.Info("Received a call")

		checkr, err := http.Get(checkURL)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(checkr.StatusCode)
	}
}

func handleCheck(appLogger *zap.SugaredLogger) func(http.ResponseWriter, *http.Request) {
	return func(
		w http.ResponseWriter, r *http.Request) {
		appLogger.Info("Received a call")
		w.WriteHeader(http.StatusOK)
	}
}