// AngelaMos | 2026
// handler.go

package user

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/carterperez-dev/templates/go-backend/internal/core"
	"github.com/carterperez-dev/templates/go-backend/internal/middleware"
)

type Handler struct {
	service   *Service
	validator *validator.Validate
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service:   service,
		validator: validator.New(validator.WithRequiredStructEnabled()),
	}
}

func (h *Handler) RegisterRoutes(
	r chi.Router,
	authenticator func(http.Handler) http.Handler,
) {
	r.Route("/users", func(r chi.Router) {
		r.Use(authenticator)

		r.Get("/me", h.GetMe)
		r.Put("/me", h.UpdateMe)
		r.Delete("/me", h.DeleteMe)
	})
}

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	user, err := h.service.GetMe(r.Context(), userID)
	if err != nil {
		if errors.Is(err, core.ErrNotFound) {
			core.NotFound(w, "user")
			return
		}
		core.InternalServerError(w, err)
		return
	}

	core.OK(w, ToUserResponse(user))
}

func (h *Handler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.BadRequest(w, "invalid request body")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		core.BadRequest(w, core.FormatValidationError(err))
		return
	}

	user, err := h.service.UpdateMe(r.Context(), userID, req)
	if err != nil {
		if errors.Is(err, core.ErrNotFound) {
			core.NotFound(w, "user")
			return
		}
		core.InternalServerError(w, err)
		return
	}

	core.OK(w, ToUserResponse(user))
}

func (h *Handler) DeleteMe(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	if err := h.service.DeleteMe(r.Context(), userID); err != nil {
		if errors.Is(err, core.ErrNotFound) {
			core.NotFound(w, "user")
			return
		}
		core.InternalServerError(w, err)
		return
	}

	core.NoContent(w)
}

// RegisterAdminRoutes registers admin-only user management endpoints.
func (h *Handler) RegisterAdminRoutes(
	r chi.Router,
	authenticator, adminOnly func(http.Handler) http.Handler,
) {
	r.Route("/admin/users", func(r chi.Router) {
		r.Use(authenticator)
		r.Use(adminOnly)

		r.Get("/", h.ListUsers)
		r.Get("/{userID}", h.GetUser)
		r.Put("/{userID}", h.UpdateUser)
		r.Put("/{userID}/role", h.UpdateUserRole)
		r.Put("/{userID}/tier", h.UpdateUserTier)
		r.Delete("/{userID}", h.DeleteUser)
	})
}

// ListUsers returns a paginated list of users with optional filtering.
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	params := ListUsersParams{
		Page:     parseIntQuery(r, "page", 1),
		PageSize: parseIntQuery(r, "page_size", 20),
		Search:   r.URL.Query().Get("search"),
		Role:     r.URL.Query().Get("role"),
		Tier:     r.URL.Query().Get("tier"),
	}

	users, total, err := h.service.ListUsers(r.Context(), params)
	if err != nil {
		core.InternalServerError(w, err)
		return
	}

	core.Paginated(
		w,
		ToUserResponseList(users),
		params.Page,
		params.PageSize,
		total,
	)
}

// GetUser returns a specific user by ID (admin only).
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")

	user, err := h.service.GetUser(r.Context(), userID)
	if err != nil {
		if errors.Is(err, core.ErrNotFound) {
			core.NotFound(w, "user")
			return
		}
		core.InternalServerError(w, err)
		return
	}

	core.OK(w, ToUserResponse(user))
}

// UpdateUser updates a specific user's profile (admin only).
func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.BadRequest(w, "invalid request body")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		core.BadRequest(w, core.FormatValidationError(err))
		return
	}

	user, err := h.service.UpdateUser(r.Context(), userID, req)
	if err != nil {
		if errors.Is(err, core.ErrNotFound) {
			core.NotFound(w, "user")
			return
		}
		core.InternalServerError(w, err)
		return
	}

	core.OK(w, ToUserResponse(user))
}

// UpdateUserRole changes a user's role (admin only).
func (h *Handler) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")

	var req UpdateUserRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.BadRequest(w, "invalid request body")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		core.BadRequest(w, core.FormatValidationError(err))
		return
	}

	user, err := h.service.UpdateUserRole(r.Context(), userID, req.Role)
	if err != nil {
		if errors.Is(err, core.ErrNotFound) {
			core.NotFound(w, "user")
			return
		}
		core.InternalServerError(w, err)
		return
	}

	core.OK(w, ToUserResponse(user))
}

// UpdateUserTier changes a user's subscription tier (admin only).
func (h *Handler) UpdateUserTier(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")

	var req UpdateUserTierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.BadRequest(w, "invalid request body")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		core.BadRequest(w, core.FormatValidationError(err))
		return
	}

	user, err := h.service.UpdateUserTier(r.Context(), userID, req.Tier)
	if err != nil {
		if errors.Is(err, core.ErrNotFound) {
			core.NotFound(w, "user")
			return
		}
		core.InternalServerError(w, err)
		return
	}

	core.OK(w, ToUserResponse(user))
}

// DeleteUser soft deletes a user account (admin only).
func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	requesterID := middleware.GetUserID(r.Context())
	targetID := chi.URLParam(r, "userID")

	if err := h.service.CanDeleteUser(r.Context(), requesterID, targetID); err != nil {
		if errors.Is(err, core.ErrForbidden) {
			core.Forbidden(w, "insufficient permissions")
			return
		}
		if errors.Is(err, core.ErrNotFound) {
			core.NotFound(w, "user")
			return
		}
		core.InternalServerError(w, err)
		return
	}

	if err := h.service.DeleteUser(r.Context(), targetID); err != nil {
		if errors.Is(err, core.ErrNotFound) {
			core.NotFound(w, "user")
			return
		}
		core.InternalServerError(w, err)
		return
	}

	core.NoContent(w)
}

func parseIntQuery(r *http.Request, key string, defaultVal int) int {
	val := r.URL.Query().Get(key)
	if val == "" {
		return defaultVal
	}

	parsed, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}

	return parsed
}
