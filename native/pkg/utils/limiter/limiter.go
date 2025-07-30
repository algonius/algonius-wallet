package limiter

import (
	"net/http"

	"golang.org/x/time/rate"
)

type RateLimiter struct {
	limiter   *rate.Limiter
	transport http.RoundTripper
}

type Option func(*RateLimiter)

func WithTransport(transport http.RoundTripper) Option {
	return func(r *RateLimiter) {
		r.transport = transport
	}
}

func NewRateLimiter(r rate.Limit, burst int, opts ...Option) *RateLimiter {
	rl := &RateLimiter{
		limiter:   rate.NewLimiter(r, burst),
		transport: http.DefaultTransport,
	}

	for _, opt := range opts {
		opt(rl)
	}

	return rl
}

func (r *RateLimiter) RoundTrip(req *http.Request) (*http.Response, error) {
	err := r.limiter.Wait(req.Context())
	if err != nil {
		return nil, err
	}

	return r.transport.RoundTrip(req)
}