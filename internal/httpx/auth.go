package httpx

import (
	"crypto/subtle"
	"net/http"
)

// AuthMiddleware validates the X-Internal-Key header using constant-time comparison.
func AuthMiddleware(expectedKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("X-Internal-Key")
			if key == "" || subtle.ConstantTimeCompare([]byte(key), []byte(expectedKey)) != 1 {
				WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or invalid X-Internal-Key header.", nil)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
