package middleware

import (
	"net/http"
	"sync"
	"time"
)

type ipEntry struct {
	count    int
	resetAt  time.Time
	mu       sync.Mutex
}

var (
	ipMap   = make(map[string]*ipEntry)
	ipMapMu sync.Mutex
)

// RateLimit returns a middleware that limits requests per minute per IP.
func RateLimit(reqPerMinute int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr

			ipMapMu.Lock()
			entry, ok := ipMap[ip]
			if !ok {
				entry = &ipEntry{resetAt: time.Now().Add(time.Minute)}
				ipMap[ip] = entry
			}
			ipMapMu.Unlock()

			entry.mu.Lock()
			if time.Now().After(entry.resetAt) {
				entry.count = 0
				entry.resetAt = time.Now().Add(time.Minute)
			}
			entry.count++
			count := entry.count
			entry.mu.Unlock()

			if count > reqPerMinute {
				http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
