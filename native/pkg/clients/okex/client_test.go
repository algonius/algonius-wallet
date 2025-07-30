package okex

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/algonius/algonius-wallet/native/pkg/config"
)

func TestNewClient(t *testing.T) {
	logger := zap.NewNop()
	
	// Test creating a client without credentials (should work but be in mock mode)
	cfg := &config.OKExConfig{
		BaseURL: "https://www.okx.com",
		Timeout: 30,
		RateLimit: &config.RateLimitConfig{
			RPM:   20,
			Burst: 1,
		},
	}
	
	client := Init(cfg, logger)
	if client == nil {
		t.Fatal("Expected client to be created")
	}
}

func TestBroadcastTransactionValidation(t *testing.T) {
	logger := zap.NewNop()
	
	cfg := &config.OKExConfig{
		BaseURL: "https://www.okx.com",
		Timeout: 30,
	}
	
	client := Init(cfg, logger)
	
	// Test invalid parameters
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	invalidParams := BroadcastTransactionParams{
		// Missing required fields
	}
	
	_, err := client.BroadcastTransaction(ctx, invalidParams)
	if err == nil {
		t.Error("Expected validation error for invalid parameters")
	}
}

func TestQueryOrdersValidation(t *testing.T) {
	logger := zap.NewNop()
	
	cfg := &config.OKExConfig{
		BaseURL: "https://www.okx.com",
		Timeout: 30,
	}
	
	client := Init(cfg, logger)
	
	// Test invalid parameters
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	invalidParams := QueryOrdersParams{
		// Missing required fields (address or accountId)
	}
	
	_, err := client.GetOrders(ctx, invalidParams)
	if err == nil {
		t.Error("Expected validation error for invalid parameters")
	}
}