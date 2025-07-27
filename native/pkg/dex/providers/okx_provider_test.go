package providers

import (
	"context"
	"testing"

	"github.com/algonius/algonius-wallet/native/pkg/dex"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestOKXProvider_GetName(t *testing.T) {
	logger := zap.NewNop()
	provider := NewOKXProvider(OKXConfig{}, logger)
	
	assert.Equal(t, "OKX", provider.GetName())
}

func TestOKXProvider_IsSupported(t *testing.T) {
	logger := zap.NewNop()
	provider := NewOKXProvider(OKXConfig{}, logger)

	tests := []struct {
		name     string
		chainID  string
		expected bool
	}{
		{"Ethereum", "1", true},
		{"BSC", "56", true},
		{"Polygon", "137", true},
		{"Solana", "501", true},
		{"Unsupported", "999", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.IsSupported(tt.chainID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOKXProvider_GetQuote_InvalidParams(t *testing.T) {
	logger := zap.NewNop()
	provider := NewOKXProvider(OKXConfig{}, logger)
	ctx := context.Background()

	// Test with invalid parameters
	params := dex.SwapParams{
		FromToken: "", // Missing from token
		ToToken:   "0xB8c77482e45F1F44dE1745F52C74426C631bDD52",
		Amount:    "1000000000000000000",
		ChainID:   "1",
	}

	quote, err := provider.GetQuote(ctx, params)
	assert.Error(t, err)
	assert.Nil(t, quote)
	assert.Contains(t, err.Error(), "invalid swap parameters")
}

func TestOKXProvider_EstimateGas_InvalidParams(t *testing.T) {
	logger := zap.NewNop()
	provider := NewOKXProvider(OKXConfig{}, logger)
	ctx := context.Background()

	// Test with invalid parameters that will fail quote fetch
	params := dex.SwapParams{
		FromToken: "", // Missing from token
		ToToken:   "0xB8c77482e45F1F44dE1745F52C74426C631bDD52",
		Amount:    "1000000000000000000",
		ChainID:   "1",
	}

	gasLimit, gasPrice, err := provider.EstimateGas(ctx, params)
	assert.Error(t, err)
	assert.Equal(t, uint64(0), gasLimit)
	assert.Equal(t, "", gasPrice)
}

func TestOKXProvider_GetBalance(t *testing.T) {
	logger := zap.NewNop()
	provider := NewOKXProvider(OKXConfig{}, logger)
	ctx := context.Background()

	// OKX provider doesn't support balance queries
	balance, err := provider.GetBalance(ctx, "0x123", "0x456", "1")
	assert.Error(t, err)
	assert.NotNil(t, balance) // Returns placeholder balance info
	assert.Contains(t, err.Error(), "balance queries not supported")
}

func TestOKXProvider_Configuration(t *testing.T) {
	logger := zap.NewNop()
	
	// Test with custom configuration
	config := OKXConfig{
		APIKey:     "test-api-key",
		SecretKey:  "test-secret",
		Passphrase: "test-passphrase",
		BaseURL:    "https://test.okx.com",
	}
	
	provider := NewOKXProvider(config, logger)
	
	assert.Equal(t, "OKX", provider.GetName())
	assert.Equal(t, config.APIKey, provider.apiKey)
	assert.Equal(t, config.SecretKey, provider.secretKey)
	assert.Equal(t, config.Passphrase, provider.passphrase)
	assert.Equal(t, config.BaseURL, provider.baseURL)
}

func TestOKXProvider_DefaultConfiguration(t *testing.T) {
	logger := zap.NewNop()
	
	// Test with default configuration
	provider := NewOKXProvider(OKXConfig{}, logger)
	
	assert.Equal(t, "https://www.okx.com", provider.baseURL)
	assert.NotNil(t, provider.httpClient)
}

func TestOKXProvider_GenerateSignature(t *testing.T) {
	logger := zap.NewNop()
	config := OKXConfig{
		SecretKey: "test-secret-key",
	}
	provider := NewOKXProvider(config, logger)
	
	// Test signature generation
	timestamp := "2023-01-01T00:00:00.000Z"
	method := "GET"
	requestPath := "/api/v5/dex/aggregator/quote"
	body := ""
	
	signature := provider.generateSignature(timestamp, method, requestPath, body)
	assert.NotEmpty(t, signature)
	
	// Test with empty secret key
	providerNoSecret := NewOKXProvider(OKXConfig{}, logger)
	emptySignature := providerNoSecret.generateSignature(timestamp, method, requestPath, body)
	assert.Empty(t, emptySignature)
}