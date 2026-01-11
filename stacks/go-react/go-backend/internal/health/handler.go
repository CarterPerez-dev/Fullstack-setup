// AngelaMos | 2026
// handler.go

package health

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-chi/chi/v5"
)

type Checker interface {
	Ping(ctx context.Context) error
}

type Handler struct {
	db       Checker
	redis    Checker
	ready    atomic.Bool
	shutdown atomic.Bool
}

func NewHandler(db, redis Checker) *Handler {
	h := &Handler{
		db:    db,
		redis: redis,
	}
	h.ready.Store(true)
	return h
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/healthz", h.Liveness)
	r.Get("/livez", h.Liveness)
	r.Get("/readyz", h.Readiness)
}

func (h *Handler) Liveness(w http.ResponseWriter, r *http.Request) {
	if h.shutdown.Load() {
		h.writeStatus(w, http.StatusServiceUnavailable, StatusResponse{
			Status: "shutting_down",
		})
		return
	}

	h.writeStatus(w, http.StatusOK, StatusResponse{
		Status: "ok",
	})
}

func (h *Handler) Readiness(w http.ResponseWriter, r *http.Request) {
	if h.shutdown.Load() {
		h.writeStatus(w, http.StatusServiceUnavailable, StatusResponse{
			Status: "shutting_down",
		})
		return
	}

	if !h.ready.Load() {
		h.writeStatus(w, http.StatusServiceUnavailable, StatusResponse{
			Status: "not_ready",
		})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	checks := h.runHealthChecks(ctx)

	allHealthy := true
	for _, check := range checks {
		if !check.Healthy {
			allHealthy = false
			break
		}
	}

	status := "ok"
	statusCode := http.StatusOK
	if !allHealthy {
		status = "degraded"
		statusCode = http.StatusServiceUnavailable
	}

	h.writeStatus(w, statusCode, ReadinessResponse{
		Status: status,
		Checks: checks,
	})
}

func (h *Handler) runHealthChecks(ctx context.Context) []HealthCheck {
	var wg sync.WaitGroup
	checks := make([]HealthCheck, 2)

	wg.Add(2)

	go func() {
		defer wg.Done()
		checks[0] = h.checkDatabase(ctx)
	}()

	go func() {
		defer wg.Done()
		checks[1] = h.checkRedis(ctx)
	}()

	wg.Wait()
	return checks
}

func (h *Handler) checkDatabase(ctx context.Context) HealthCheck {
	check := HealthCheck{
		Name:    "database",
		Healthy: true,
	}

	if h.db == nil {
		check.Healthy = false
		check.Message = "database checker not configured"
		return check
	}

	start := time.Now()
	err := h.db.Ping(ctx)
	check.Latency = time.Since(start).String()

	if err != nil {
		check.Healthy = false
		check.Message = "ping failed"
	}

	return check
}

func (h *Handler) checkRedis(ctx context.Context) HealthCheck {
	check := HealthCheck{
		Name:    "redis",
		Healthy: true,
	}

	if h.redis == nil {
		check.Healthy = false
		check.Message = "redis checker not configured"
		return check
	}

	start := time.Now()
	err := h.redis.Ping(ctx)
	check.Latency = time.Since(start).String()

	if err != nil {
		check.Healthy = false
		check.Message = "ping failed"
	}

	return check
}

func (h *Handler) SetReady(ready bool) {
	h.ready.Store(ready)
}

func (h *Handler) SetShutdown(shutdown bool) {
	h.shutdown.Store(shutdown)
}

func (h *Handler) writeStatus(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.WriteHeader(status)
	//nolint:errcheck // best-effort response
	_ = json.NewEncoder(w).Encode(data)
}

type StatusResponse struct {
	Status string `json:"status"`
}

type ReadinessResponse struct {
	Status string        `json:"status"`
	Checks []HealthCheck `json:"checks"`
}

type HealthCheck struct {
	Name    string `json:"name"`
	Healthy bool   `json:"healthy"`
	Latency string `json:"latency,omitempty"`
	Message string `json:"message,omitempty"`
}
