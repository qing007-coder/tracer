package collector

import (
	"context"
	"golang.org/x/time/rate"
)

type RateLimiter struct {
	limiter *rate.Limiter
}

func NewRateLimiter() *RateLimiter {
	r := new(RateLimiter)
	r.init()

	return r
}

func (r *RateLimiter) init() {
	r.limiter = rate.NewLimiter(100, 200)
}

func (r *RateLimiter) Allow() bool {
	return r.limiter.Allow()
}

func (r *RateLimiter) Wait() error {
	return r.limiter.Wait(context.Background())
}
