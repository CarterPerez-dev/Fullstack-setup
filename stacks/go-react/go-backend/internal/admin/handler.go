// AngelaMos | 2026
// handler.go

package admin

import (
	"context"
	"database/sql"
	"net/http"
	"runtime"

	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"

	"github.com/carterperez-dev/templates/go-backend/internal/core"
)

type AuthService interface {
	InvalidateAllSessions(ctx context.Context) error
}

type Handler struct {
	dbStats    func() sql.DBStats
	redisStats func() *redis.PoolStats
	redisPing  func(ctx context.Context) error
	dbPing     func(ctx context.Context) error
	authSvc    AuthService
}

type HandlerConfig struct {
	DBStats    func() sql.DBStats
	RedisStats func() *redis.PoolStats
	RedisPing  func(ctx context.Context) error
	DBPing     func(ctx context.Context) error
	AuthSvc    AuthService
}

func NewHandler(cfg HandlerConfig) *Handler {
	return &Handler{
		dbStats:    cfg.DBStats,
		redisStats: cfg.RedisStats,
		redisPing:  cfg.RedisPing,
		dbPing:     cfg.DBPing,
		authSvc:    cfg.AuthSvc,
	}
}

func (h *Handler) RegisterRoutes(
	r chi.Router,
	authenticator, adminOnly func(http.Handler) http.Handler,
) {
	r.Route("/admin", func(r chi.Router) {
		r.Use(authenticator)
		r.Use(adminOnly)

		r.Get("/stats", h.GetSystemStats)
		r.Get("/stats/db", h.GetDatabaseStats)
		r.Get("/stats/redis", h.GetRedisStats)
		r.Get("/stats/runtime", h.GetRuntimeStats)
	})
}

func (h *Handler) GetSystemStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	dbHealthy := true
	if h.dbPing != nil {
		if err := h.dbPing(ctx); err != nil {
			dbHealthy = false
		}
	}

	redisHealthy := true
	if h.redisPing != nil {
		if err := h.redisPing(ctx); err != nil {
			redisHealthy = false
		}
	}

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	response := SystemStatsResponse{
		Database: DatabaseStatus{
			Healthy: dbHealthy,
			Stats:   h.getDBStats(),
		},
		Redis: RedisStatus{
			Healthy: redisHealthy,
			Stats:   h.getRedisStats(),
		},
		Runtime: RuntimeStats{
			GoVersion:    runtime.Version(),
			NumGoroutine: runtime.NumGoroutine(),
			NumCPU:       runtime.NumCPU(),
			MemAlloc:     memStats.Alloc,
			MemSys:       memStats.Sys,
			NumGC:        memStats.NumGC,
		},
	}

	core.OK(w, response)
}

func (h *Handler) GetDatabaseStats(w http.ResponseWriter, r *http.Request) {
	core.OK(w, h.getDBStats())
}

func (h *Handler) GetRedisStats(w http.ResponseWriter, r *http.Request) {
	core.OK(w, h.getRedisStats())
}

func (h *Handler) GetRuntimeStats(w http.ResponseWriter, r *http.Request) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	response := RuntimeStats{
		GoVersion:    runtime.Version(),
		NumGoroutine: runtime.NumGoroutine(),
		NumCPU:       runtime.NumCPU(),
		MemAlloc:     memStats.Alloc,
		MemSys:       memStats.Sys,
		NumGC:        memStats.NumGC,
	}

	core.OK(w, response)
}

func (h *Handler) getDBStats() *DBPoolStats {
	if h.dbStats == nil {
		return nil
	}

	stats := h.dbStats()
	return &DBPoolStats{
		MaxOpenConnections: stats.MaxOpenConnections,
		OpenConnections:    stats.OpenConnections,
		InUse:              stats.InUse,
		Idle:               stats.Idle,
		WaitCount:          stats.WaitCount,
		WaitDuration:       stats.WaitDuration.String(),
		MaxIdleClosed:      stats.MaxIdleClosed,
		MaxIdleTimeClosed:  stats.MaxIdleTimeClosed,
		MaxLifetimeClosed:  stats.MaxLifetimeClosed,
	}
}

func (h *Handler) getRedisStats() *RedisPoolStats {
	if h.redisStats == nil {
		return nil
	}

	stats := h.redisStats()
	return &RedisPoolStats{
		Hits:       stats.Hits,
		Misses:     stats.Misses,
		Timeouts:   stats.Timeouts,
		TotalConns: stats.TotalConns,
		IdleConns:  stats.IdleConns,
		StaleConns: stats.StaleConns,
	}
}

type SystemStatsResponse struct {
	Database DatabaseStatus `json:"database"`
	Redis    RedisStatus    `json:"redis"`
	Runtime  RuntimeStats   `json:"runtime"`
}

type DatabaseStatus struct {
	Healthy bool         `json:"healthy"`
	Stats   *DBPoolStats `json:"stats,omitempty"`
}

type RedisStatus struct {
	Healthy bool            `json:"healthy"`
	Stats   *RedisPoolStats `json:"stats,omitempty"`
}

type DBPoolStats struct {
	MaxOpenConnections int    `json:"max_open_connections"`
	OpenConnections    int    `json:"open_connections"`
	InUse              int    `json:"in_use"`
	Idle               int    `json:"idle"`
	WaitCount          int64  `json:"wait_count"`
	WaitDuration       string `json:"wait_duration"`
	MaxIdleClosed      int64  `json:"max_idle_closed"`
	MaxIdleTimeClosed  int64  `json:"max_idle_time_closed"`
	MaxLifetimeClosed  int64  `json:"max_lifetime_closed"`
}

type RedisPoolStats struct {
	Hits       uint32 `json:"hits"`
	Misses     uint32 `json:"misses"`
	Timeouts   uint32 `json:"timeouts"`
	TotalConns uint32 `json:"total_conns"`
	IdleConns  uint32 `json:"idle_conns"`
	StaleConns uint32 `json:"stale_conns"`
}

type RuntimeStats struct {
	GoVersion    string `json:"go_version"`
	NumGoroutine int    `json:"num_goroutine"`
	NumCPU       int    `json:"num_cpu"`
	MemAlloc     uint64 `json:"mem_alloc_bytes"`
	MemSys       uint64 `json:"mem_sys_bytes"`
	NumGC        uint32 `json:"num_gc"`
}
