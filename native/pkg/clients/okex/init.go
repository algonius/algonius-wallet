package okex

import (
	"context"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"
	"golang.org/x/time/rate"

	"github.com/algonius/algonius-wallet/native/pkg/config"
	"github.com/algonius/algonius-wallet/native/pkg/utils/limiter"
)

// Init creates an OKEx client from configuration
func Init(cfg *config.OKExConfig, logger *zap.Logger) IOKEXClient {
	if cfg == nil {
		logger.Warn("OKEx config is nil, creating client with defaults")
		cfg = &config.OKExConfig{
			BaseURL: "https://www.okx.com",
			Timeout: 30,
		}
	}

	baseURL := cfg.BaseURL
	projectID := cfg.ProjectID
	apiKey := cfg.APIKey
	apiSecret := cfg.SecretKey
	passphrase := cfg.PassPhrase
	timeout := time.Duration(cfg.Timeout) * time.Second

	// Allow environment variables to override config values
	if okexApiKey := os.Getenv("OKEX_API_KEY"); okexApiKey != "" {
		apiKey = okexApiKey
	}

	if okexApiSecret := os.Getenv("OKEX_API_SECRET"); okexApiSecret != "" {
		apiSecret = okexApiSecret
	}

	if okexPassphrase := os.Getenv("OKEX_PASSPHRASE"); okexPassphrase != "" {
		passphrase = okexPassphrase
	}

	if okexProjectID := os.Getenv("OKEX_PROJECT_ID"); okexProjectID != "" {
		projectID = okexProjectID
	}

	var transport http.RoundTripper
	if cfg.RateLimit != nil && cfg.RateLimit.RPM > 0 {
		rpm := cfg.RateLimit.RPM
		burst := cfg.RateLimit.Burst
		if burst == 0 {
			burst = rpm
		}
		interval := time.Minute / time.Duration(rpm)
		transport = limiter.NewRateLimiter(rate.Every(interval), burst)
	}

	options := []ClientOption{WithTimeout(timeout), WithLogger(logger)}
	if transport != nil {
		options = append(options, WithTransport(transport))
	}

	client := NewClient(baseURL, projectID, apiKey, apiSecret, passphrase, options...)

	// Test connection by querying supported chains (only if credentials are provided)
	if apiKey != "" && apiSecret != "" && passphrase != "" {
		chains, err := client.GetSupportedChains(context.Background(), "")
		if err != nil {
			logger.Warn("Failed to query OKEx supported chains during initialization", 
				zap.Error(err))
		} else if len(chains.Data) > 0 {
			logger.Info("OKEx client initialized successfully",
				zap.Int("supported_chains", len(chains.Data)))
		}
	} else {
		logger.Debug("OKEx client initialized without credentials (mock mode)")
	}

	return client
}