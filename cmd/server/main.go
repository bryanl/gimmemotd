package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

type spec struct {
	Addr int `default:"8181"`
}

func main() {
	logger := logrus.New()

	var s spec
	err := envconfig.Process("gimmemotd", &s)
	if err != nil {
		logger.WithError(err).Fatal("unable to process environment")
	}

	logger.Info("server is starting")

	router := http.NewServeMux()
	router.Handle("/", index(logger))

	addr := fmt.Sprintf(":%d", s.Addr)

	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  10 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		logger.Println("shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			logger.WithError(err).Fatal("unable to gracefully shutdown server")
		}
	}()

	logger.WithField("addr", addr).Info("server is ready")
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		logger.WithError(err).
			WithField("addr", addr).
			Fatal("server could not listen on address")
	}

	logger.Info("server stopped")
}

type response struct {
	Hostname string `json:"hostname,omitempty"`
	Message  string `json:"message,omitempty"`
}

func index(logger logrus.FieldLogger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()

		defer func() {
			logger.WithField("elapsed", time.Since(now)).Info("index")
		}()

		resp := response{
			Hostname: "host",
			Message:  "v2",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(&resp)
	})
}
