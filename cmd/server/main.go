package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ankit-lilly/go-datastar-daisyui-template/internal/config"
	"github.com/ankit-lilly/go-datastar-daisyui-template/internal/handlers"
	"github.com/ankit-lilly/go-datastar-daisyui-template/internal/jobs"
)

func main() {
	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load configuration
	cfg := config.Load()

	// Initialize job hub for background tasks
	jobHub := jobs.NewHub(logger)
	go jobHub.Run()

	// Setup routes
	mux := http.NewServeMux()
	h := handlers.New(logger, jobHub)

	// Static files
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Pages
	mux.HandleFunc("GET /", h.Index)

	// API - SSE endpoints
	mux.HandleFunc("GET /api/counter", h.Counter)
	mux.HandleFunc("POST /api/increment", h.Increment)
	mux.HandleFunc("POST /api/job/start", h.StartJob)

	// Create server
	server := &http.Server{
		Addr:         cfg.Addr,
		Handler:      logRequests(logger, mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 0, // Disabled for SSE
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("server starting", "addr", cfg.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop job hub
	jobHub.Stop()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("server stopped gracefully")
}

func logRequests(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logger.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"duration", time.Since(start),
			"remote", r.RemoteAddr,
		)
	})
}
