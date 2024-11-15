package webserver

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tberk-s/learning-url-shortener-with-go/src/internal/config"
	"github.com/tberk-s/learning-url-shortener-with-go/src/internal/db"
	"github.com/tberk-s/learning-url-shortener-with-go/src/internal/transport/http/urlshortenerhandler"
)

const (
	DefaultReadTimeout     = 10 * time.Second
	DefaultWriteTimeout    = 10 * time.Second
	DefaultIdleTimeout     = 15 * time.Second
	DefaultShutdownTimeout = 10 * time.Second
)

// WebServer represents the web server instance
type WebServer struct {
	config *config.Config
	db     *db.DB
}

type Option func(*WebServer)

func WithDBName(name string) Option {
	return func(s *WebServer) {
		s.config.DBName = name
	}
}

func WithDBHost(host string) Option {
	return func(s *WebServer) {
		s.config.DBHost = host
	}
}

func WithDBUser(user string) Option {
	return func(s *WebServer) {
		s.config.DBUser = user
	}
}

func WithDBPassword(password string) Option {
	return func(s *WebServer) {
		s.config.DBPassword = password
	}
}

func WithServerEnv(env string) Option {
	return func(s *WebServer) {
		s.config.ServerEnv = env
	}
}

func WithDBPort(port int) Option {
	return func(s *WebServer) {
		s.config.DBPort = port
	}
}

func New(opts ...Option) error {
	ws := &WebServer{
		config: config.LoadConfig(),
	}
	for _, opt := range opts {
		opt(ws)
	}
	if ws.config.ServerEnv == "" {
		ws.config.ServerEnv = "development"
	}

	database, err := db.New(
		ws.config.DBUser,
		ws.config.DBPassword,
		ws.config.DBHost,
		ws.config.DBName,
		ws.config.DBPort,
	)

	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	ws.db = database

	urlHandler, err := urlshortenerhandler.New(ws.db)
	if err != nil {
		return fmt.Errorf("failed to create URL handler: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/shorten", urlHandler.ShowShortenPage())
	mux.HandleFunc("/home", urlshortenerhandler.ShowHomePage) // Move home page to explicit path
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/home", http.StatusPermanentRedirect)
		} else {
			urlshortenerhandler.RedirectHandler(ws.db)(w, r)
		}
	})

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
		log.Println("Starting API server on", webServer.Addr)
		if err := webServer.ListenAndServe(); err != nil {
			webError <- err
		}
	}()

	select {
	case err := <-webError:
		log.Printf("Server error: %v", err)
		ws.db.Close()

		return err

	case sig := <-shutdown:
		log.Printf("Received shutdown signal: %s", sig)

		ctx, cancel := context.WithTimeout(context.Background(), DefaultShutdownTimeout)
		defer cancel()

		if err := webServer.Shutdown(ctx); err != nil {
			log.Printf("Graceful shutdown failed, forcing server close: %v", err)
		}

		ws.db.Close() // Close DB after successful shutdown
		if ctx.Err() != nil {
			log.Printf("Shutdown timed out: %v", ctx.Err())
		}
		log.Println("Server shutdown gracefully")
	}

	return nil
}
