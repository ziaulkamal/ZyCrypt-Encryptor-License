package middleware

import (
	"net/http"
)

// HMACAuth is a placeholder for request-level HMAC verification.
// Actual HMAC verification is done inside each handler since it requires body parsing.
// This middleware can be used for header-level pre-checks if needed in the future.
func HMACAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}
}
