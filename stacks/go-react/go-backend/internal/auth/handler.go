// AngelaMos | 2026
// handler.go

package auth

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strings"

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
	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", h.Login)
		r.Post("/register", h.Register)
		r.Post("/refresh", h.Refresh)

		r.Group(func(r chi.Router) {
			r.Use(authenticator)
			r.Get("/me", h.GetMe)
			r.Post("/logout", h.Logout)
			r.Post("/logout-all", h.LogoutAll)
			r.Get("/sessions", h.GetSessions)
			r.Delete("/sessions/{sessionID}", h.RevokeSession)
			r.Post("/change-password", h.ChangePassword)
		})
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.BadRequest(w, "invalid request body")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		core.BadRequest(w, core.FormatValidationError(err))
		return
	}

	userAgent := r.UserAgent()
	ipAddress := extractIPAddress(r)

	resp, err := h.service.Login(r.Context(), req, userAgent, ipAddress)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			core.JSONError(
				w,
				core.UnauthorizedError("invalid email or password"),
			)
			return
		}
		core.InternalServerError(w, err)
		return
	}

	core.OK(w, resp)
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.BadRequest(w, "invalid request body")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		core.BadRequest(w, core.FormatValidationError(err))
		return
	}

	userAgent := r.UserAgent()
	ipAddress := extractIPAddress(r)

	resp, err := h.service.Register(r.Context(), req, userAgent, ipAddress)
	if err != nil {
		if errors.Is(err, ErrEmailExists) {
			core.JSONError(w, core.DuplicateError("email"))
			return
		}
		core.InternalServerError(w, err)
		return
	}

	core.Created(w, resp)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.BadRequest(w, "invalid request body")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		core.BadRequest(w, core.FormatValidationError(err))
		return
	}

	userAgent := r.UserAgent()
	ipAddress := extractIPAddress(r)

	resp, err := h.service.Refresh(
		r.Context(),
		req.RefreshToken,
		userAgent,
		ipAddress,
	)
	if err != nil {
		if errors.Is(err, ErrTokenReuse) {
			core.JSONError(w, core.NewAppError(
				core.ErrTokenRevoked,
				"security alert: token reuse detected, all sessions revoked",
				http.StatusUnauthorized,
				"TOKEN_REUSE_DETECTED",
			))
			return
		}
		if errors.Is(err, core.ErrTokenExpired) {
			core.JSONError(w, core.TokenExpiredError())
			return
		}
		if errors.Is(err, core.ErrTokenRevoked) {
			core.JSONError(w, core.TokenRevokedError())
			return
		}
		if errors.Is(err, core.ErrTokenInvalid) {
			core.JSONError(w, core.TokenInvalidError())
			return
		}
		core.InternalServerError(w, err)
		return
	}

	core.OK(w, resp)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		core.Unauthorized(w, "")
		return
	}

	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.BadRequest(w, "invalid request body")
		return
	}

	if err := h.service.Logout(r.Context(), req.RefreshToken, userID); err != nil {
		if errors.Is(err, core.ErrForbidden) {
			core.Forbidden(w, "cannot revoke another user's token")
			return
		}
		core.InternalServerError(w, err)
		return
	}

	core.NoContent(w)
}

func (h *Handler) LogoutAll(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		core.Unauthorized(w, "")
		return
	}

	if err := h.service.LogoutAll(r.Context(), userID); err != nil {
		core.InternalServerError(w, err)
		return
	}

	core.NoContent(w)
}

func (h *Handler) GetSessions(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		core.Unauthorized(w, "")
		return
	}

	sessions, err := h.service.GetActiveSessions(r.Context(), userID)
	if err != nil {
		core.InternalServerError(w, err)
		return
	}

	core.OK(w, SessionsResponse{Sessions: sessions})
}

func (h *Handler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		core.Unauthorized(w, "")
		return
	}

	sessionID := chi.URLParam(r, "sessionID")
	if sessionID == "" {
		core.BadRequest(w, "session ID required")
		return
	}

	if err := h.service.RevokeSession(r.Context(), userID, sessionID); err != nil {
		if errors.Is(err, core.ErrNotFound) {
			core.NotFound(w, "session")
			return
		}
		if errors.Is(err, core.ErrForbidden) {
			core.Forbidden(w, "cannot revoke another user's session")
			return
		}
		core.InternalServerError(w, err)
		return
	}

	core.NoContent(w)
}

func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		core.Unauthorized(w, "")
		return
	}

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.BadRequest(w, "invalid request body")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		core.BadRequest(w, core.FormatValidationError(err))
		return
	}

	if err := h.service.ChangePassword(r.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			core.JSONError(
				w,
				core.UnauthorizedError("current password is incorrect"),
			)
			return
		}
		core.InternalServerError(w, err)
		return
	}

	core.NoContent(w)
}

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		core.Unauthorized(w, "")
		return
	}

	user, err := h.service.GetCurrentUser(r.Context(), userID)
	if err != nil {
		if errors.Is(err, core.ErrNotFound) {
			core.NotFound(w, "user")
			return
		}
		core.InternalServerError(w, err)
		return
	}

	core.OK(w, user)
}

func extractIPAddress(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[len(ips)-1])
	}

	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}
