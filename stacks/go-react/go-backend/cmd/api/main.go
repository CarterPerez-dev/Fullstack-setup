// AngelaMos | 2026
// main.go

package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/carterperez-dev/templates/go-backend/internal/admin"
	"github.com/carterperez-dev/templates/go-backend/internal/auth"
	"github.com/carterperez-dev/templates/go-backend/internal/config"
	"github.com/carterperez-dev/templates/go-backend/internal/core"
	"github.com/carterperez-dev/templates/go-backend/internal/health"
	"github.com/carterperez-dev/templates/go-backend/internal/middleware"
	"github.com/carterperez-dev/templates/go-backend/internal/server"
	"github.com/carterperez-dev/templates/go-backend/internal/user"
)

const (
	drainDelay = 5 * time.Second
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	if err := run(*configPath); err != nil {
		slog.Error("application error", "error", err)
		os.Exit(1)
	}
}

//nolint:funlen // bootstrap code is inherently verbose
func run(configPath string) error {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}

	logger := setupLogger(cfg.Log)
	slog.SetDefault(logger)

	logger.Info("starting application",
		"name", cfg.App.Name,
		"version", cfg.App.Version,
		"environment", cfg.App.Environment,
	)

	var telemetry *core.Telemetry
	if cfg.Otel.Enabled {
		tel, telErr := core.NewTelemetry(ctx, cfg.Otel, cfg.App)
		if telErr != nil {
			logger.Warn("failed to initialize telemetry", "error", telErr)
		} else {
			telemetry = tel
			logger.Info("OpenTelemetry tracer initialized",
				"endpoint", cfg.Otel.Endpoint,
			)
		}
	}

	db, err := core.NewDatabase(ctx, cfg.Database)
	if err != nil {
		return err
	}
	logger.Info("database connected",
		"max_open_conns", cfg.Database.MaxOpenConns,
		"max_idle_conns", cfg.Database.MaxIdleConns,
	)

	redis, err := core.NewRedis(ctx, cfg.Redis)
	if err != nil {
		return err
	}
	logger.Info("redis connected",
		"pool_size", cfg.Redis.PoolSize,
	)

	jwtManager, err := auth.NewJWTManager(cfg.JWT)
	if err != nil {
		return err
	}
	logger.Info("JWT manager initialized",
		"algorithm", "ES256",
		"key_id", jwtManager.GetKeyID(),
	)

	userRepo := user.NewRepository(db.DB)
	userSvc := user.NewService(userRepo)
	userHandler := user.NewHandler(userSvc)

	authRepo := auth.NewRepository(db.DB)
	authSvc := auth.NewService(authRepo, jwtManager, userSvc, redis.Client)
	authHandler := auth.NewHandler(authSvc)

	healthHandler := health.NewHandler(db, redis)

	adminHandler := admin.NewHandler(admin.HandlerConfig{
		DBStats:    db.Stats,
		RedisStats: redis.PoolStats,
		DBPing:     db.Ping,
		RedisPing:  redis.Ping,
	})

	srv := server.New(server.Config{
		ServerConfig:  cfg.Server,
		HealthHandler: healthHandler,
		Logger:        logger,
	})

	router := srv.Router()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger(logger))
	router.Use(
		middleware.NewRateLimiter(redis.Client, middleware.RateLimitConfig{
			Limit: middleware.PerMinute(
				cfg.RateLimit.Requests,
				cfg.RateLimit.Burst,
			),
			FailOpen: true,
		}).Handler,
	)
	router.Use(middleware.SecurityHeaders(cfg.App.Environment == "production"))
	router.Use(middleware.CORS(cfg.CORS))

	healthHandler.RegisterRoutes(router)

	router.Get("/.well-known/jwks.json", jwtManager.GetJWKSHandler())

	authenticator := middleware.Authenticator(jwtManager)
	adminOnly := middleware.RequireAdmin

	router.Route("/v1", func(r chi.Router) {
		authHandler.RegisterRoutes(r, authenticator)

		r.Post("/users", authHandler.Register)

		userHandler.RegisterRoutes(r, authenticator)
		userHandler.RegisterAdminRoutes(r, authenticator, adminOnly)
		adminHandler.RegisterRoutes(r, authenticator, adminOnly)
	})

	errChan := make(chan error, 1)
	go func() {
		errChan <- srv.Start()
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		cfg.Server.ShutdownTimeout+drainDelay+5*time.Second,
	)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx, drainDelay); err != nil {
		logger.Error("server shutdown error", "error", err)
	}

	if telemetry != nil {
		if err := telemetry.Shutdown(shutdownCtx); err != nil {
			logger.Error("telemetry shutdown error", "error", err)
		}
	}

	if err := redis.Close(); err != nil {
		logger.Error("redis close error", "error", err)
	}

	if err := db.Close(); err != nil {
		logger.Error("database close error", "error", err)
	}

	logger.Info("application stopped")
	return nil
}

func setupLogger(cfg config.LogConfig) *slog.Logger {
	var handler slog.Handler

	level := slog.LevelInfo
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	opts := &slog.HandlerOptions{Level: level}

	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}
