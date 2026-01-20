package collector

import (
	"context"
	"golang.org/x/time/rate"
)

// RateLimiter limits the rate of operations using a token bucket algorithm.
type RateLimiter struct {
	limiter *rate.Limiter
}

// NewRateLimiter creates a new RateLimiter.
func NewRateLimiter() *RateLimiter {
	r := new(RateLimiter)
	r.init()

	return r
}

// init initializes the rate limiter with a limit of 100 events/sec and a burst of 200.
func (r *RateLimiter) init() {
	r.limiter = rate.NewLimiter(100, 200)
}

// Allow checks if a new event is allowed to happen immediately.
func (r *RateLimiter) Allow() bool {
	return r.limiter.Allow()
}

// Wait blocks until an event is allowed to happen.
func (r *RateLimiter) Wait() error {
	return r.limiter.Wait(context.Background())
}
