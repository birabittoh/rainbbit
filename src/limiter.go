package src

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type limiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	limiters    = make(map[string]*limiterEntry)
	limiterMu   sync.Mutex
	limiterOnce sync.Once
	// globalLimiter protects the server from overall high load.
	// Adjust these values based on expected traffic.
	globalLimiter = rate.NewLimiter(rate.Limit(100), 200)
)

func cleanupLimiters() {
	for {
		time.Sleep(1 * time.Hour)
		limiterMu.Lock()
		for ip, entry := range limiters {
			if time.Since(entry.lastSeen) > 1*time.Hour {
				delete(limiters, ip)
			}
		}
		limiterMu.Unlock()
	}
}

func rateLimiterMiddleware(next http.Handler) http.Handler {
	limiterOnce.Do(func() {
		go cleanupLimiters()
	})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Global rate limit
		if !globalLimiter.Allow() {
			http.Error(w, "Server busy", http.StatusServiceUnavailable)
			return
		}

		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, "Can't find IP address", http.StatusInternalServerError)
			return
		}

		limiterMu.Lock()
		entry, exists := limiters[ip]
		if !exists {
			entry = &limiterEntry{
				limiter: rate.NewLimiter(3, 30), // 3 richieste al secondo, burst massimo di 30
			}
			limiters[ip] = entry
		}
		entry.lastSeen = time.Now()
		limiter := entry.limiter
		limiterMu.Unlock()

		if !limiter.Allow() {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
