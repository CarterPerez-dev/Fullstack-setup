// AngelaMos | 2026
// auth.go

package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/carterperez-dev/templates/go-backend/internal/core"
)

const (
	UserIDKey   contextKey = "user_id"
	UserRoleKey contextKey = "user_role"
	UserTierKey contextKey = "user_tier"
	ClaimsKey   contextKey = "jwt_claims"
)

type TokenVerifier interface {
	VerifyAccessToken(
		ctx context.Context,
		token string,
	) (*AccessTokenClaims, error)
}

type AccessTokenClaims struct {
	UserID       string
	Role         string
	Tier         string
	TokenVersion int
}

func Authenticator(verifier TokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := ExtractToken(r)

			if token == "" {
				core.JSONError(
					w,
					core.UnauthorizedError("missing authorization token"),
				)
				return
			}

			claims, err := verifier.VerifyAccessToken(r.Context(), token)
			if err != nil {
				handleAuthError(w, err)
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserRoleKey, claims.Role)
			ctx = context.WithValue(ctx, UserTierKey, claims.Tier)
			ctx = context.WithValue(ctx, ClaimsKey, claims)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func OptionalAuth(verifier TokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := ExtractToken(r)

			if token != "" {
				claims, err := verifier.VerifyAccessToken(r.Context(), token)
				if err == nil {
					ctx := r.Context()
					ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
					ctx = context.WithValue(ctx, UserRoleKey, claims.Role)
					ctx = context.WithValue(ctx, UserTierKey, claims.Tier)
					ctx = context.WithValue(ctx, ClaimsKey, claims)
					r = r.WithContext(ctx)
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RequireRole(roles ...string) func(http.Handler) http.Handler {
	roleSet := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		roleSet[role] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole := GetUserRole(r.Context())

			if userRole == "" {
				core.JSONError(
					w,
					core.UnauthorizedError("authentication required"),
				)
				return
			}

			if _, ok := roleSet[userRole]; !ok {
				core.JSONError(
					w,
					core.ForbiddenError("insufficient permissions"),
				)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RequireAdmin(next http.Handler) http.Handler {
	return RequireRole("admin")(next)
}

func ExtractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return ""
	}

	return strings.TrimSpace(parts[1])
}

func handleAuthError(w http.ResponseWriter, err error) {
	if core.IsAppError(err) {
		core.JSONError(w, err)
		return
	}

	switch {
	case errors.Is(err, core.ErrTokenExpired):
		core.JSONError(w, core.TokenExpiredError())
	case errors.Is(err, core.ErrTokenRevoked):
		core.JSONError(w, core.TokenRevokedError())
	case errors.Is(err, core.ErrTokenInvalid):
		core.JSONError(w, core.TokenInvalidError())
	default:
		core.JSONError(w, core.TokenInvalidError())
	}
}

func GetUserID(ctx context.Context) string {
	if id, ok := ctx.Value(UserIDKey).(string); ok {
		return id
	}
	return ""
}

func GetUserRole(ctx context.Context) string {
	if role, ok := ctx.Value(UserRoleKey).(string); ok {
		return role
	}
	return ""
}

func GetUserTier(ctx context.Context) string {
	if tier, ok := ctx.Value(UserTierKey).(string); ok {
		return tier
	}
	return ""
}

func GetClaims(ctx context.Context) *AccessTokenClaims {
	if claims, ok := ctx.Value(ClaimsKey).(*AccessTokenClaims); ok {
		return claims
	}
	return nil
}

func IsAuthenticated(ctx context.Context) bool {
	return GetUserID(ctx) != ""
}

func IsAdmin(ctx context.Context) bool {
	return GetUserRole(ctx) == "admin"
}
