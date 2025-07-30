package jito

import (
	"testing"

	"github.com/algonius/algonius-wallet/native/pkg/config"
)

func TestNewJitoClient(t *testing.T) {
	// Test creating a Jito client without API key (should work but be in mock mode)
	cfg := &config.JitoConfig{
		Enabled:         false,
		BaseURL:         "https://mainnet.block-engine.jito.wtf/api/v1",
		APIKey:          "",
		BaseTipLamports: 1000,
		MaxTipLamports:  100000,
		TipStrategy:     "exponential",
	}
	
	client := Init(cfg)
	if client == nil {
		t.Fatal("Expected Jito client to be created")
	}
}

func TestJitoClientWithDefaults(t *testing.T) {
	// Test with nil config (should use defaults)
	client := Init(nil)
	if client == nil {
		t.Fatal("Expected Jito client to be created with defaults")
	}
}