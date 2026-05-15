// internal/middleware/recovery.go
package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"runtime/debug"
)

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				logRecoveredPanic(r.Context(), r, rec)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"errors":[{"message":"internal server error"}]}`))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func logRecoveredPanic(ctx context.Context, r *http.Request, recovered any) {
	slog.LogAttrs(
		ctx,
		slog.LevelError,
		"http panic recovered",
		slog.String("component", "http"),
		slog.String("event", "panic"),
		slog.String("request_id", requestIDFromHeader(r)),
		slog.String("http_method", r.Method),
		slog.String("http_path", r.URL.Path),
		slog.Int("http_status", http.StatusInternalServerError),
		slog.Any("panic", recovered),
		slog.String("stack", string(debug.Stack())),
	)
}
