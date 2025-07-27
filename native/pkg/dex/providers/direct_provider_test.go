// SPDX-License-Identifier: Apache-2.0
package providers

import (
	"context"
	"testing"

	"github.com/algonius/algonius-wallet/native/pkg/dex"
	"go.uber.org/zap/zaptest"
)

func TestDirectProvider_GetName(t *testing.T) {
	logger := zaptest.NewLogger(t)
	provider := NewDirectProvider(logger)
	
	if provider.GetName() != "Direct" {
		t.Errorf("Expected provider name 'Direct', got '%s'", provider.GetName())
	}
}

func TestDirectProvider_IsSupported(t *testing.T) {
	logger := zaptest.NewLogger(t)
	provider := NewDirectProvider(logger)
	
	tests := []struct {
		chainID   string
		supported bool
	}{
		{"1", true},   // Ethereum
		{"56", true},  // BSC
		{"501", true}, // Solana
		{"137", false}, // Polygon (not supported)
		{"", false},   // Empty
	}
	
	for _, tt := range tests {
		t.Run(tt.chainID, func(t *testing.T) {
			supported := provider.IsSupported(tt.chainID)
			if supported != tt.supported {
				t.Errorf("Expected IsSupported(%s) = %v, got %v", tt.chainID, tt.supported, supported)
			}
		})
	}
}

func TestDirectProvider_GetQuote(t *testing.T) {
	logger := zaptest.NewLogger(t)
	provider := NewDirectProvider(logger)
	ctx := context.Background()
	
	params := dex.SwapParams{
		FromToken:   "ETH",
		ToToken:     "USDT",
		Amount:      "1.0",
		Slippage:    0.005,
		FromAddress: "0x742d35Cc6673C4C5f9aB9e3Be0A78a19a4B43c89",
		ToAddress:   "0x742d35Cc6673C4C5f9aB9e3Be0A78a19a4B43c89",
		ChainID:     "1",
		PrivateKey:  "0x1234567890abcdef",
	}
	
	quote, err := provider.GetQuote(ctx, params)
	if err != nil {
		t.Errorf("Failed to get quote: %v", err)
	}
	
	if quote.FromToken != params.FromToken {
		t.Errorf("Expected FromToken %s, got %s", params.FromToken, quote.FromToken)
	}
	
	if quote.ToToken != params.ToToken {
		t.Errorf("Expected ToToken %s, got %s", params.ToToken, quote.ToToken)
	}
	
	if quote.Provider != "Direct" {
		t.Errorf("Expected Provider 'Direct', got '%s'", quote.Provider)
	}
	
	if quote.Slippage != params.Slippage {
		t.Errorf("Expected Slippage %f, got %f", params.Slippage, quote.Slippage)
	}
}

func TestDirectProvider_GetQuote_InvalidParams(t *testing.T) {
	logger := zaptest.NewLogger(t)
	provider := NewDirectProvider(logger)
	ctx := context.Background()
	
	tests := []struct {
		name   string
		params dex.SwapParams
	}{
		{
			name: "missing from token",
			params: dex.SwapParams{
				ToToken:     "USDT",
				Amount:      "1.0",
				Slippage:    0.005,
				FromAddress: "0x742d35Cc6673C4C5f9aB9e3Be0A78a19a4B43c89",
				ChainID:     "1",
			},
		},
		{
			name: "missing to token",
			params: dex.SwapParams{
				FromToken:   "ETH",
				Amount:      "1.0",
				Slippage:    0.005,
				FromAddress: "0x742d35Cc6673C4C5f9aB9e3Be0A78a19a4B43c89",
				ChainID:     "1",
			},
		},
		{
			name: "missing amount",
			params: dex.SwapParams{
				FromToken:   "ETH",
				ToToken:     "USDT",
				Slippage:    0.005,
				FromAddress: "0x742d35Cc6673C4C5f9aB9e3Be0A78a19a4B43c89",
				ChainID:     "1",
			},
		},
		{
			name: "unsupported chain",
			params: dex.SwapParams{
				FromToken:   "ETH",
				ToToken:     "USDT",
				Amount:      "1.0",
				Slippage:    0.005,
				FromAddress: "0x742d35Cc6673C4C5f9aB9e3Be0A78a19a4B43c89",
				ChainID:     "999", // Unsupported
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := provider.GetQuote(ctx, tt.params)
			if err == nil {
				t.Errorf("Expected error for %s, but got none", tt.name)
			}
		})
	}
}

func TestDirectProvider_ExecuteSwap(t *testing.T) {
	logger := zaptest.NewLogger(t)
	provider := NewDirectProvider(logger)
	ctx := context.Background()
	
	params := dex.SwapParams{
		FromToken:   "ETH",
		ToToken:     "USDT",
		Amount:      "1.0",
		Slippage:    0.005,
		FromAddress: "0x742d35Cc6673C4C5f9aB9e3Be0A78a19a4B43c89",
		ToAddress:   "0x742d35Cc6673C4C5f9aB9e3Be0A78a19a4B43c89",
		ChainID:     "1",
		PrivateKey:  "0x1234567890abcdef",
	}
	
	result, err := provider.ExecuteSwap(ctx, params)
	if err != nil {
		t.Errorf("Failed to execute swap: %v", err)
	}
	
	if result.FromToken != params.FromToken {
		t.Errorf("Expected FromToken %s, got %s", params.FromToken, result.FromToken)
	}
	
	if result.ToToken != params.ToToken {
		t.Errorf("Expected ToToken %s, got %s", params.ToToken, result.ToToken)
	}
	
	if result.Provider != "Direct" {
		t.Errorf("Expected Provider 'Direct', got '%s'", result.Provider)
	}
	
	if result.Status != "pending" {
		t.Errorf("Expected Status 'pending', got '%s'", result.Status)
	}
	
	// Check transaction hash format based on chain
	if params.ChainID == "1" || params.ChainID == "56" {
		// EVM chains should have 0x prefix and 64 characters
		if len(result.TxHash) != 66 || result.TxHash[:2] != "0x" {
			t.Errorf("Invalid EVM transaction hash format: %s", result.TxHash)
		}
	} else if params.ChainID == "501" {
		// Solana should have 64 characters without 0x prefix
		if len(result.TxHash) != 64 {
			t.Errorf("Invalid Solana transaction hash format: %s", result.TxHash)
		}
	}
}

func TestDirectProvider_GetBalance(t *testing.T) {
	logger := zaptest.NewLogger(t)
	provider := NewDirectProvider(logger)
	ctx := context.Background()
	
	tests := []struct {
		chainID      string
		address      string
		tokenAddress string
		expectedSymbol string
		expectedDecimals int
	}{
		{
			chainID:      "1",
			address:      "0x742d35Cc6673C4C5f9aB9e3Be0A78a19a4B43c89",
			tokenAddress: "",
			expectedSymbol: "ETH",
			expectedDecimals: 18,
		},
		{
			chainID:      "1",
			address:      "0x742d35Cc6673C4C5f9aB9e3Be0A78a19a4B43c89",
			tokenAddress: "ETH",
			expectedSymbol: "ETH",
			expectedDecimals: 18,
		},
		{
			chainID:      "56",
			address:      "0x742d35Cc6673C4C5f9aB9e3Be0A78a19a4B43c89",
			tokenAddress: "",
			expectedSymbol: "BNB",
			expectedDecimals: 18,
		},
		{
			chainID:      "501",
			address:      "9WzDXwBbmkg8ZTbNMqUxvQRAyrZzDsGYdLVL9zYtAWWM",
			tokenAddress: "",
			expectedSymbol: "SOL",
			expectedDecimals: 9,
		},
		{
			chainID:      "1",
			address:      "0x742d35Cc6673C4C5f9aB9e3Be0A78a19a4B43c89",
			tokenAddress: "0xA0b86a33E6417aAA9f79dF12E43d04Be87da7C9c",
			expectedSymbol: "TOKEN",
			expectedDecimals: 18,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.chainID+"_"+tt.expectedSymbol, func(t *testing.T) {
			balance, err := provider.GetBalance(ctx, tt.address, tt.tokenAddress, tt.chainID)
			if err != nil {
				t.Errorf("Failed to get balance: %v", err)
			}
			
			if balance.TokenSymbol != tt.expectedSymbol {
				t.Errorf("Expected TokenSymbol %s, got %s", tt.expectedSymbol, balance.TokenSymbol)
			}
			
			if balance.Decimals != tt.expectedDecimals {
				t.Errorf("Expected Decimals %d, got %d", tt.expectedDecimals, balance.Decimals)
			}
			
			if balance.Balance == "" {
				t.Error("Expected non-empty balance")
			}
		})
	}
}

func TestDirectProvider_EstimateGas(t *testing.T) {
	logger := zaptest.NewLogger(t)
	provider := NewDirectProvider(logger)
	ctx := context.Background()
	
	tests := []struct {
		chainID        string
		expectedGasLimit uint64
		expectedGasPrice string
	}{
		{"1", 150000, "20"},   // Ethereum
		{"56", 120000, "5"},   // BSC
		{"501", 5000, "1"},    // Solana
	}
	
	for _, tt := range tests {
		t.Run(tt.chainID, func(t *testing.T) {
			params := dex.SwapParams{
				ChainID: tt.chainID,
			}
			
			gasLimit, gasPrice, err := provider.EstimateGas(ctx, params)
			if err != nil {
				t.Errorf("Failed to estimate gas: %v", err)
			}
			
			if gasLimit != tt.expectedGasLimit {
				t.Errorf("Expected gas limit %d, got %d", tt.expectedGasLimit, gasLimit)
			}
			
			if gasPrice != tt.expectedGasPrice {
				t.Errorf("Expected gas price %s, got %s", tt.expectedGasPrice, gasPrice)
			}
		})
	}
}

func TestDirectProvider_EstimateGas_UnsupportedChain(t *testing.T) {
	logger := zaptest.NewLogger(t)
	provider := NewDirectProvider(logger)
	ctx := context.Background()
	
	params := dex.SwapParams{
		ChainID: "999", // Unsupported chain
	}
	
	_, _, err := provider.EstimateGas(ctx, params)
	if err == nil {
		t.Error("Expected error for unsupported chain")
	}
}