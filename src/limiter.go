package src

import (
	"net"
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

func rateLimiterMiddleware(next http.Handler) http.Handler {
	// Mappa degli IP con il rispettivo rate limiter
	limiters := make(map[string]*rate.Limiter)
	var limiterMutex = &sync.Mutex{}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, "Can't find IP address", http.StatusInternalServerError)
			return
		}

		limiterMutex.Lock()
		limiter, exists := limiters[ip]
		if !exists {
			limiter = rate.NewLimiter(1, 15) // 1 richiesta al secondo, burst massimo di 15
			limiters[ip] = limiter
		}
		limiterMutex.Unlock()

		if !limiter.Allow() {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
