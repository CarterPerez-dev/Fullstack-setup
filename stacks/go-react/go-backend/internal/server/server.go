// AngelaMos | 2026
// server.go

package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/carterperez-dev/templates/go-backend/internal/config"
	"github.com/carterperez-dev/templates/go-backend/internal/health"
)

type Server struct {
	httpServer    *http.Server
	router        *chi.Mux
	config        config.ServerConfig
	healthHandler *health.Handler
	logger        *slog.Logger
}

type Config struct {
	ServerConfig  config.ServerConfig
	HealthHandler *health.Handler
	Logger        *slog.Logger
}

func New(cfg Config) *Server {
	router := chi.NewRouter()

	router.Use(chimw.CleanPath)
	router.Use(chimw.StripSlashes)

	return &Server{
		httpServer: &http.Server{
			Addr:         cfg.ServerConfig.Address(),
			Handler:      router,
			ReadTimeout:  cfg.ServerConfig.ReadTimeout,
			WriteTimeout: cfg.ServerConfig.WriteTimeout,
			IdleTimeout:  cfg.ServerConfig.IdleTimeout,
		},
		router:        router,
		config:        cfg.ServerConfig,
		healthHandler: cfg.HealthHandler,
		logger:        cfg.Logger,
	}
}

func (s *Server) Router() *chi.Mux {
	return s.router
}

func (s *Server) Start() error {
	s.logger.Info("starting HTTP server",
		"addr", s.config.Address(),
		"read_timeout", s.config.ReadTimeout,
		"write_timeout", s.config.WriteTimeout,
		"idle_timeout", s.config.IdleTimeout,
	)

	if err := s.httpServer.ListenAndServe(); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("http server error: %w", err)
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context, drainDelay time.Duration) error {
	s.logger.Info("initiating graceful shutdown")

	s.logger.Info("marking server as not ready")
	if s.healthHandler != nil {
		s.healthHandler.SetReady(false)
		s.healthHandler.SetShutdown(true)
	}

	s.logger.Info("waiting for load balancer to drain",
		"delay", drainDelay,
	)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(drainDelay):
	}

	s.logger.Info("stopping HTTP server")
	shutdownCtx, cancel := context.WithTimeout(ctx, s.config.ShutdownTimeout)
	defer cancel()

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("http server shutdown: %w", err)
	}

	s.logger.Info("HTTP server stopped gracefully")
	return nil
}

func (s *Server) Address() string {
	return s.httpServer.Addr
}
