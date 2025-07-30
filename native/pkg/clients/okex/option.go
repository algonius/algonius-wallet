package okex

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// ClientOption defines the type for client options
type ClientOption func(*Client)

// WithTimeout sets custom timeout for HTTP client
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.client.Timeout = timeout
	}
}

// WithTransport sets custom transport for HTTP client
func WithTransport(transport http.RoundTripper) ClientOption {
	return func(c *Client) {
		c.client.Transport = transport
	}
}

// WithLogger sets custom logger for HTTP client
func WithLogger(logger *zap.Logger) ClientOption {
	return func(c *Client) {
		c.logger = logger
	}
}