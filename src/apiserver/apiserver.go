package apiserver

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tberk-s/learning-url-shortener-with-go/src/internal/controllers"
)

const (
	DefaultReadTimeout     = 10 * time.Second
	DefaultWriteTimeout    = 10 * time.Second
	DefaultIdleTimeout     = 15 * time.Second
	DefaultShutdownTimeout = 10 * time.Second
)

type WebServer struct {
	serverEnv string
}

type Option func(*WebServer)

func WithServerEnv(env string) Option {
	return func(s *WebServer) {
		s.serverEnv = env
	}
}

func New(opts ...Option) error {
	ws := &WebServer{}
	for _, opt := range opts {
		opt(ws)
	}
	if ws.serverEnv == "" {
		ws.serverEnv = "development"
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", controllers.ShowHomePage)
	mux.HandleFunc("/shorten", controllers.ShowShortenPage)

	webServer := &http.Server{
		Addr:         ":8000",
		Handler:      mux,
		ReadTimeout:  DefaultReadTimeout,
		WriteTimeout: DefaultWriteTimeout,
		IdleTimeout:  DefaultIdleTimeout,
	}
	shutdown := make(chan os.Signal, 1)
	webError := make(chan error)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Println("Starting web server on port 8000")
		if err := webServer.ListenAndServe(); err != nil {
			log.Fatalf("Failed to start web server: %v", err)
		}
	}()

	select {
	case err := <-webError:

		return err
	case sig := <-shutdown:
		log.Printf("Received shutdown signal: %s", sig)

		ctx, cancel := context.WithTimeout(context.Background(), DefaultShutdownTimeout)
		defer cancel()

		if err := webServer.Shutdown(ctx); err != nil {
			log.Printf("Graceful shutdown failed, forcing server close: %v", err)

			return fmt.Errorf("error during server shutdown: %w", err)
		}
		log.Println("Server shutdown gracefully")
	}

	return nil
}
