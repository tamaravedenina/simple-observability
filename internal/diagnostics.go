package internal

import (
	"go.uber.org/zap"
	"net"
	"net/http"

	"github.com/gorilla/mux"
)

// Diagnostics responsible for diagnostics logic of the app
func Diagnostics(port string, appLogger *zap.SugaredLogger, shutdown chan<- error) *http.Server {
	r := mux.NewRouter()
	r.HandleFunc("/health", handleHealth())

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

func handleHealth() func(http.ResponseWriter, *http.Request) {
	return func(
		w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}