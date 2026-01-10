// AngelaMos | 2026
// headers.go

package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/carterperez-dev/templates/go-backend/internal/config"
)

func SecurityHeaders(isProduction bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()

			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("X-Frame-Options", "DENY")
			h.Set("X-XSS-Protection", "1; mode=block")
			h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
			h.Set(
				"Permissions-Policy",
				"geolocation=(), microphone=(), camera=()",
			)

			if isProduction {
				h.Set(
					"Strict-Transport-Security",
					"max-age=31536000; includeSubDomains; preload",
				)
			}

			h.Set("Content-Security-Policy", buildCSP(isProduction))

			next.ServeHTTP(w, r)
		})
	}
}

func buildCSP(isProduction bool) string {
	directives := []string{
		"default-src 'self'",
		"script-src 'self'",
		"style-src 'self' 'unsafe-inline'",
		"img-src 'self' data: https:",
		"font-src 'self'",
		"connect-src 'self'",
		"frame-ancestors 'none'",
		"base-uri 'self'",
		"form-action 'self'",
	}

	if !isProduction {
		directives[1] = "script-src 'self' 'unsafe-inline' 'unsafe-eval'"
	}

	return strings.Join(directives, "; ")
}

func CORS(cfg config.CORSConfig) func(http.Handler) http.Handler {
	allowedOrigins := make(map[string]struct{}, len(cfg.AllowedOrigins))
	for _, origin := range cfg.AllowedOrigins {
		allowedOrigins[origin] = struct{}{}
	}

	methodsStr := strings.Join(cfg.AllowedMethods, ", ")
	headersStr := strings.Join(cfg.AllowedHeaders, ", ")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if origin != "" {
				if _, ok := allowedOrigins[origin]; ok {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Vary", "Origin")

					if cfg.AllowCredentials {
						w.Header().
							Set("Access-Control-Allow-Credentials", "true")
					}
				}
			}

			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Methods", methodsStr)
				w.Header().Set("Access-Control-Allow-Headers", headersStr)

				if cfg.MaxAge > 0 {
					w.Header().
						Set("Access-Control-Max-Age", strconv.Itoa(cfg.MaxAge))
				}

				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
