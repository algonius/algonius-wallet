// SPDX-License-Identifier: Apache-2.0
package chain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenMappingConfig_GetTokenInfo(t *testing.T) {
	config := NewTokenMappingConfig()

	tests := []struct {
		name             string
		tokenIdentifier  string
		expectedSymbol   string
		expectedChain    string
		expectedNative   bool
		expectedDecimals int
		expectError      bool
	}{
		{
			name:             "ETH primary symbol",
			tokenIdentifier:  "ETH",
			expectedSymbol:   "ETH",
			expectedChain:    "ETH",
			expectedNative:   true,
			expectedDecimals: 18,
			expectError:      false,
		},
		{
			name:             "ETH alias - ETHER",
			tokenIdentifier:  "ETHER",
			expectedSymbol:   "ETH",
			expectedChain:    "ETH",
			expectedNative:   true,
			expectedDecimals: 18,
			expectError:      false,
		},
		{
			name:             "BNB primary symbol",
			tokenIdentifier:  "BNB",
			expectedSymbol:   "BNB",
			expectedChain:    "BSC",
			expectedNative:   true,
			expectedDecimals: 18,
			expectError:      false,
		},
		{
			name:             "BNB alias - BINANCE",
			tokenIdentifier:  "BINANCE",
			expectedSymbol:   "BNB",
			expectedChain:    "BSC",
			expectedNative:   true,
			expectedDecimals: 18,
			expectError:      false,
		},
		{
			name:             "SOL primary symbol",
			tokenIdentifier:  "SOL",
			expectedSymbol:   "SOL",
			expectedChain:    "SOL",
			expectedNative:   true,
			expectedDecimals: 9,
			expectError:      false,
		},
		{
			name:             "SOL alias - SOLANA",
			tokenIdentifier:  "SOLANA",
			expectedSymbol:   "SOL",
			expectedChain:    "SOL",
			expectedNative:   true,
			expectedDecimals: 9,
			expectError:      false,
		},
		{
			name:            "Unsupported token",
			tokenIdentifier: "UNSUPPORTED_TOKEN",
			expectError:     true,
		},
		{
			name:             "Empty token defaults to ETH",
			tokenIdentifier:  "",
			expectedSymbol:   "ETH",
			expectedChain:    "ETH",
			expectedNative:   true,
			expectedDecimals: 18,
			expectError:      false,
		},
		{
			name:             "Case insensitive - lowercase",
			tokenIdentifier:  "eth",
			expectedSymbol:   "ETH",
			expectedChain:    "ETH",
			expectedNative:   true,
			expectedDecimals: 18,
			expectError:      false,
		},
		{
			name:             "Case insensitive - mixed case",
			tokenIdentifier:  "BnB",
			expectedSymbol:   "BNB",
			expectedChain:    "BSC",
			expectedNative:   true,
			expectedDecimals: 18,
			expectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenInfo, err := config.GetTokenInfo(tt.tokenIdentifier)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, tokenInfo)
			} else {
				require.NoError(t, err)
				require.NotNil(t, tokenInfo)
				assert.Equal(t, tt.expectedSymbol, tokenInfo.Symbol)
				assert.Equal(t, tt.expectedChain, tokenInfo.ChainName)
				assert.Equal(t, tt.expectedNative, tokenInfo.IsNative)
				assert.Equal(t, tt.expectedDecimals, tokenInfo.Decimals)
			}
		})
	}
}

func TestTokenMappingConfig_GetChainForToken(t *testing.T) {
	config := NewTokenMappingConfig()

	tests := []struct {
		name            string
		tokenIdentifier string
		expectedChain   string
		expectError     bool
	}{
		{"ETH to ETH chain", "ETH", "ETH", false},
		{"BNB to BSC chain", "BNB", "BSC", false},
		{"SOL to SOL chain", "SOL", "SOL", false},
		{"ETHER alias to ETH chain", "ETHER", "ETH", false},
		{"BINANCE alias to BSC chain", "BINANCE", "BSC", false},
		{"Unsupported token", "UNKNOWN", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chain, err := config.GetChainForToken(tt.tokenIdentifier)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, chain)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedChain, chain)
			}
		})
	}
}

func TestTokenMappingConfig_IsNativeToken(t *testing.T) {
	config := NewTokenMappingConfig()

	tests := []struct {
		name            string
		tokenIdentifier string
		expectedNative  bool
		expectError     bool
	}{
		{"ETH is native", "ETH", true, false},
		{"BNB is native", "BNB", true, false},
		{"SOL is native", "SOL", true, false},
		{"ETHER alias is native", "ETHER", true, false},
		{"Unsupported token", "UNKNOWN", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isNative, err := config.IsNativeToken(tt.tokenIdentifier)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedNative, isNative)
			}
		})
	}
}

func TestTokenMappingConfig_ValidateTokenForChain(t *testing.T) {
	config := NewTokenMappingConfig()

	tests := []struct {
		name            string
		tokenIdentifier string
		chainName       string
		expectError     bool
	}{
		{"ETH on ETH chain", "ETH", "ETH", false},
		{"BNB on BSC chain", "BNB", "BSC", false},
		{"SOL on SOL chain", "SOL", "SOL", false},
		{"ETH on wrong chain", "ETH", "BSC", true},
		{"BNB on wrong chain", "BNB", "ETH", true},
		{"SOL on wrong chain", "SOL", "ETH", true},
		{"Case insensitive chain", "ETH", "eth", false},
		{"Unsupported token", "UNKNOWN", "ETH", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := config.ValidateTokenForChain(tt.tokenIdentifier, tt.chainName)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTokenMappingConfig_GetSupportedTokens(t *testing.T) {
	config := NewTokenMappingConfig()

	supportedTokens := config.GetSupportedTokens()

	// Should contain all primary symbols and aliases
	expectedTokens := []string{"ETH", "ETHER", "ETHEREUM", "BNB", "BINANCE", "BINANCE_COIN", "SOL", "SOLANA"}

	assert.Len(t, supportedTokens, len(expectedTokens))

	for _, expected := range expectedTokens {
		assert.Contains(t, supportedTokens, expected)
	}
}

func TestTokenMappingConfig_GetNativeTokensPerChain(t *testing.T) {
	config := NewTokenMappingConfig()

	nativeTokens := config.GetNativeTokensPerChain()

	expected := map[string]string{
		"ETH": "ETH",
		"BSC": "BNB",
		"SOL": "SOL",
	}

	assert.Equal(t, expected, nativeTokens)
}

func TestTokenMappingConfig_AddCustomToken(t *testing.T) {
	config := NewTokenMappingConfig()

	// Add a custom token
	config.AddCustomToken("TEST", []string{"TESTCOIN"}, "ETH", false, 6, "Test token")

	// Verify it was added correctly
	tokenInfo, err := config.GetTokenInfo("TEST")
	require.NoError(t, err)
	assert.Equal(t, "TEST", tokenInfo.Symbol)
	assert.Equal(t, "ETH", tokenInfo.ChainName)
	assert.False(t, tokenInfo.IsNative)
	assert.Equal(t, 6, tokenInfo.Decimals)

	// Verify alias works
	aliasInfo, err := config.GetTokenInfo("TESTCOIN")
	require.NoError(t, err)
	assert.Equal(t, tokenInfo, aliasInfo)
}

func TestDefaultTokenMapping(t *testing.T) {
	// Test that the global instance is properly initialized
	assert.NotNil(t, DefaultTokenMapping)

	// Test basic functionality
	tokenInfo, err := DefaultTokenMapping.GetTokenInfo("ETH")
	require.NoError(t, err)
	assert.Equal(t, "ETH", tokenInfo.Symbol)
	assert.Equal(t, "ETH", tokenInfo.ChainName)
	assert.True(t, tokenInfo.IsNative)
}