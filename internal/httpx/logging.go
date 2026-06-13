package httpx

import (
	"log/slog"
	"net/http"
	"time"
)

// LoggingMiddleware logs each request with trace_id and latency.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"trace_id", r.Header.Get("X-Trace-ID"),
			"latency_ms", time.Since(start).Milliseconds(),
		)
	})
}
