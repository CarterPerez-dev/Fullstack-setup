// AngelaMos | 2026
// logging.go

package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/trace"
)

type loggerKey struct{}

func Logger(baseLogger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			requestID := GetRequestID(r.Context())

			reqLogger := baseLogger.With(
				slog.String("request_id", requestID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
			)

			if span := trace.SpanFromContext(r.Context()); span.SpanContext().
				IsValid() {
				reqLogger = reqLogger.With(
					slog.String(
						"trace_id",
						span.SpanContext().TraceID().String(),
					),
					slog.String(
						"span_id",
						span.SpanContext().SpanID().String(),
					),
				)
			}

			ctx := context.WithValue(r.Context(), loggerKey{}, reqLogger)

			ww := &responseWriter{
				ResponseWriter: w,
				status:         http.StatusOK,
			}

			next.ServeHTTP(ww, r.WithContext(ctx))

			latency := time.Since(start)

			logLevel := slog.LevelInfo
			if ww.status >= 500 {
				logLevel = slog.LevelError
			} else if ww.status >= 400 {
				logLevel = slog.LevelWarn
			}

			reqLogger.Log(r.Context(), logLevel, "request completed",
				slog.Int("status", ww.status),
				slog.Int("bytes", ww.bytes),
				slog.Duration("latency", latency),
				slog.String("user_agent", r.UserAgent()),
			)
		})
	}
}

func GetLogger(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}

type responseWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytes += n
	return n, err
}

func (rw *responseWriter) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}
