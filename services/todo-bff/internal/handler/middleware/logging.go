// internal/middleware/logging.go
package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := newResponseWriter(w)

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		logHTTPRequest(r.Context(), r, wrapped.statusCode, duration)
	})
}

func requestIDFromHeader(r *http.Request) string {
	for _, key := range []string{"X-Request-Id", "Request-Id", "Trace-Id"} {
		if v := r.Header.Get(key); v != "" {
			return v
		}
	}
	return "-"
}

func httpLogLevel(statusCode int) slog.Level {
	switch {
	case statusCode < 400:
		return slog.LevelInfo
	case statusCode < 500:
		return slog.LevelWarn
	default:
		return slog.LevelError
	}
}

func logHTTPRequest(
	ctx context.Context,
	r *http.Request,
	statusCode int,
	duration time.Duration,
) {
	slog.LogAttrs(
		ctx,
		httpLogLevel(statusCode),
		"http request",
		slog.String("component", "http"),
		slog.String("request_id", requestIDFromHeader(r)),
		slog.String("http_method", r.Method),
		slog.String("http_path", r.URL.Path),
		slog.Int("http_status", statusCode),
		slog.Int64("duration_ms", duration.Milliseconds()),
		slog.String("remote_addr", r.RemoteAddr),
		slog.String("user_agent", r.UserAgent()),
	)
}
