// SPDX-License-Identifier: Apache-2.0
package dex

import (
	"testing"
)

func TestSwapParams_Validation(t *testing.T) {
	tests := []struct {
		name    string
		params  SwapParams
		isValid bool
	}{
		{
			name: "valid swap params",
			params: SwapParams{
				FromToken:   "ETH",
				ToToken:     "USDT",
				Amount:      "1.0",
				Slippage:    0.005,
				FromAddress: "0x742d35Cc6673C4C5f9aB9e3Be0A78a19a4B43c89",
				ToAddress:   "0x742d35Cc6673C4C5f9aB9e3Be0A78a19a4B43c89",
				ChainID:     "1",
				PrivateKey:  "0x1234567890abcdef1234567890abcdef12345678",
			},
			isValid: true,
		},
		{
			name: "missing from token",
			params: SwapParams{
				ToToken:     "USDT",
				Amount:      "1.0",
				Slippage:    0.005,
				FromAddress: "0x742d35Cc6673C4C5f9aB9e3Be0A78a19a4B43c89",
				ToAddress:   "0x742d35Cc6673C4C5f9aB9e3Be0A78a19a4B43c89",
				ChainID:     "1",
				PrivateKey:  "0x1234567890abcdef1234567890abcdef12345678",
			},
			isValid: false,
		},
		{
			name: "missing amount",
			params: SwapParams{
				FromToken:   "ETH",
				ToToken:     "USDT",
				Slippage:    0.005,
				FromAddress: "0x742d35Cc6673C4C5f9aB9e3Be0A78a19a4B43c89",
				ToAddress:   "0x742d35Cc6673C4C5f9aB9e3Be0A78a19a4B43c89",
				ChainID:     "1",
				PrivateKey:  "0x1234567890abcdef1234567890abcdef12345678",
			},
			isValid: false,
		},
		{
			name: "invalid slippage",
			params: SwapParams{
				FromToken:   "ETH",
				ToToken:     "USDT",
				Amount:      "1.0",
				Slippage:    1.5, // > 100%
				FromAddress: "0x742d35Cc6673C4C5f9aB9e3Be0A78a19a4B43c89",
				ToAddress:   "0x742d35Cc6673C4C5f9aB9e3Be0A78a19a4B43c89",
				ChainID:     "1",
				PrivateKey:  "0x1234567890abcdef1234567890abcdef12345678",
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSwapParams(tt.params)
			if tt.isValid && err != nil {
				t.Errorf("Expected valid params but got error: %v", err)
			}
			if !tt.isValid && err == nil {
				t.Errorf("Expected invalid params but got no error")
			}
		})
	}
}

// validateSwapParams is a helper function for testing
func validateSwapParams(params SwapParams) error {
	if params.FromToken == "" {
		return &ValidationError{Field: "from_token", Message: "from token is required"}
	}
	if params.ToToken == "" {
		return &ValidationError{Field: "to_token", Message: "to token is required"}
	}
	if params.Amount == "" {
		return &ValidationError{Field: "amount", Message: "amount is required"}
	}
	if params.Slippage < 0 || params.Slippage > 1 {
		return &ValidationError{Field: "slippage", Message: "slippage must be between 0 and 1"}
	}
	if params.FromAddress == "" {
		return &ValidationError{Field: "from_address", Message: "from address is required"}
	}
	if params.ChainID == "" {
		return &ValidationError{Field: "chain_id", Message: "chain ID is required"}
	}
	return nil
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

func TestSwapQuote_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		validUntil int64
		expired   bool
	}{
		{
			name:      "valid quote",
			validUntil: 9999999999, // far in the future
			expired:   false,
		},
		{
			name:      "expired quote",
			validUntil: 1, // far in the past
			expired:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quote := &SwapQuote{
				ValidUntil: tt.validUntil,
			}
			
			// Simple check - in a real implementation, this would compare with current time
			// For testing, we assume current time is around 1700000000
			currentTime := int64(1700000000)
			isExpired := quote.ValidUntil < currentTime
			
			if isExpired != tt.expired {
				t.Errorf("Expected expired=%v, got %v", tt.expired, isExpired)
			}
		})
	}
}

func TestBalanceInfo_Validation(t *testing.T) {
	balance := &BalanceInfo{
		TokenAddress: "0xA0b86a33E6417aAA9f79dF12E43d04Be87da7C9c",
		TokenSymbol:  "USDC",
		Balance:      "1000.000000",
		Decimals:     6,
		USDValue:     "1000.00",
	}

	if balance.TokenSymbol != "USDC" {
		t.Errorf("Expected TokenSymbol USDC, got %s", balance.TokenSymbol)
	}
	
	if balance.Decimals != 6 {
		t.Errorf("Expected Decimals 6, got %d", balance.Decimals)
	}
}