package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/azizjon12/url-shortener/internal/config"
	"github.com/azizjon12/url-shortener/internal/lib/logger/handlers/slogpretty"
	"github.com/azizjon12/url-shortener/internal/lib/logger/sl"
	"github.com/azizjon12/url-shortener/internal/storage/sqlite"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/azizjon12/url-shortener/internal/http-server/handlers/redirect"
	"github.com/azizjon12/url-shortener/internal/http-server/handlers/remove"
	"github.com/azizjon12/url-shortener/internal/http-server/handlers/url/save"
	mwLogger "github.com/azizjon12/url-shortener/internal/http-server/middleware/logger"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)
	log = log.With(slog.String("env", cfg.Env))

	log.Info("initializing server", slog.String("address", cfg.Address))
	log.Debug("logger debug mode enabled")

	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error("failed to initialize storage", sl.Err(err))
	}

	_ = storage

	router := chi.NewRouter()

	router.Use(middleware.RequestID) // Add request_id every request for tracing
	router.Use(middleware.Logger)    // Logging of all requests
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer) // If there's a panic in the server, app should not close
	router.Use(middleware.URLFormat) // URL parsing for incoming requests

	router.Post("/url", save.New(log, storage))
	router.Get("/{alias}", redirect.New(log, storage))
	router.Delete("/{alias}", remove.New(log, storage))

	log.Info("starting server", slog.String("address", cfg.Address))

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTmeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server")
	}

	log.Error("server stopped")

	// TODO: run server
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()

	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)

	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
