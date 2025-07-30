package jito

import (
	"os"

	jitorpc "github.com/jito-labs/jito-go-rpc"
	"github.com/algonius/algonius-wallet/native/pkg/config"
)

// Init creates a Jito client from configuration
func Init(cfg *config.JitoConfig) IJitoAPI {
	if cfg == nil {
		cfg = &config.JitoConfig{
			Enabled:        false,
			BaseURL:       "https://mainnet.block-engine.jito.wtf/api/v1",
			APIKey:        "",
		}
	}

	baseURL := cfg.BaseURL
	apiKey := cfg.APIKey

	// Allow environment variable to override config value
	if jitoApiKey := os.Getenv("JITO_API_KEY"); jitoApiKey != "" {
		apiKey = jitoApiKey
	}

	return jitorpc.NewJitoJsonRpcClient(baseURL, apiKey)
}