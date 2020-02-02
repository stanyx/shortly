package utils

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// IPRateLimiter ...
type IPRateLimiter struct {
	*rate.Limiter
	created time.Time
}

// RateLimit ...
func RateLimit(getLimiter func(w http.ResponseWriter, r *http.Request) *rate.Limiter) func(next http.Handler) http.Handler {

	var mu sync.Mutex
	limiterStorage := make(map[string]*IPRateLimiter)

	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for k, v := range limiterStorage {
				if time.Since(v.created) > 3*time.Minute {
					delete(limiterStorage, k)
				}
			}
			mu.Unlock()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			mu.Lock()
			defer mu.Unlock()

			ip := GetIPAdress(r)

			ipLimiter, exists := limiterStorage[ip]
			if !exists {
				ipLimiter = &IPRateLimiter{getLimiter(w, r), time.Now()}
				// Include the current time when creating a new visitor.
				limiterStorage[ip] = ipLimiter
			} else {
				ipLimiter.created = time.Now()
			}

			if !ipLimiter.Allow() {
				http.Error(w, "limit exceeded", http.StatusServiceUnavailable)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
