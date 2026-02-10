package src

import (
	"log"
	"net"
	"net/http"
	"sync"

	lru "github.com/hashicorp/golang-lru/v2"
	"golang.org/x/time/rate"
)

var (
	limiterCache *lru.Cache[string, *rate.Limiter]
	limiterOnce  sync.Once
	// globalLimiter protects the server from overall high load.
	globalLimiter = rate.NewLimiter(rate.Limit(100), 200)
)

func rateLimiterMiddleware(next http.Handler) http.Handler {
	limiterOnce.Do(func() {
		var err error
		limiterCache, err = lru.New[string, *rate.Limiter](1000)
		if err != nil {
			log.Fatalf("Errore nella creazione dell'LRU cache: %v", err)
		}
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

		limiter, ok := limiterCache.Get(ip)
		if !ok {
			limiter = rate.NewLimiter(3, 30) // 3 richieste al secondo, burst massimo di 30
			limiterCache.Add(ip, limiter)
		}

		if !limiter.Allow() {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
