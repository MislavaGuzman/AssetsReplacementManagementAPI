package ratelimiter

import (
	"net/http"
	"time"
)

type Limiter interface {
	Allow(ip string) (bool, time.Duration)
}

type Config struct {
	RequestsPerTimeFrame int
	TimeFrame            time.Duration
	Enabled              bool
}

func New(requestsPerMinute int, burst int) Limiter {
	window := time.Minute
	return NewFixedWindowLimiter(requestsPerMinute, window)
}

func NewWithConfig(config Config) Limiter {
	return NewFixedWindowLimiter(config.RequestsPerTimeFrame, config.TimeFrame)
}

func (rl *FixedWindowRateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		allowed, retryAfter := rl.Allow(ip)

		if !allowed {
			w.Header().Set("Retry-After", retryAfter.String())
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
