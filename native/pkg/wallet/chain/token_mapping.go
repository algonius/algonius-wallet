// SPDX-License-Identifier: Apache-2.0
package chain

import (
	"fmt"
	"strings"
)

// TokenInfo contains metadata about a supported token
type TokenInfo struct {
	Symbol      string   `json:"symbol"`       // Primary symbol (e.g., "ETH")
	Aliases     []string `json:"aliases"`      // Alternative names (e.g., ["ETHER"])
	ChainName   string   `json:"chain_name"`   // Target chain for this token (e.g., "ETH", "BSC", "SOL")
	IsNative    bool     `json:"is_native"`    // Whether this is the native token of the chain
	Decimals    int      `json:"decimals"`     // Token decimals
	Description string   `json:"description"`  // Human-readable description
}

// TokenMappingConfig provides centralized token identifier resolution
type TokenMappingConfig struct {
	tokens map[string]*TokenInfo
}

// NewTokenMappingConfig creates a new token mapping configuration with default mappings
func NewTokenMappingConfig() *TokenMappingConfig {
	config := &TokenMappingConfig{
		tokens: make(map[string]*TokenInfo),
	}
	
	// Initialize with standard token mappings
	config.initializeDefaultMappings()
	
	return config
}

// initializeDefaultMappings sets up the standard token identifier mappings
func (tm *TokenMappingConfig) initializeDefaultMappings() {
	// Ethereum native token
	ethToken := &TokenInfo{
		Symbol:      "ETH",
		Aliases:     []string{"ETHER", "ETHEREUM"},
		ChainName:   "ETH",
		IsNative:    true,
		Decimals:    18,
		Description: "Ethereum native token",
	}
	tm.registerToken(ethToken)

	// Binance Smart Chain native token
	bnbToken := &TokenInfo{
		Symbol:      "BNB",
		Aliases:     []string{"BINANCE", "BINANCE_COIN"},
		ChainName:   "BSC",
		IsNative:    true,
		Decimals:    18,
		Description: "Binance Smart Chain native token",
	}
	tm.registerToken(bnbToken)

	// Solana native token
	solToken := &TokenInfo{
		Symbol:      "SOL",
		Aliases:     []string{"SOLANA"},
		ChainName:   "SOL",
		IsNative:    true,
		Decimals:    9,
		Description: "Solana native token",
	}
	tm.registerToken(solToken)
}

// registerToken registers a token and all its aliases in the mapping
func (tm *TokenMappingConfig) registerToken(token *TokenInfo) {
	// Register primary symbol
	normalizedSymbol := strings.ToUpper(strings.TrimSpace(token.Symbol))
	tm.tokens[normalizedSymbol] = token
	
	// Register all aliases
	for _, alias := range token.Aliases {
		normalizedAlias := strings.ToUpper(strings.TrimSpace(alias))
		tm.tokens[normalizedAlias] = token
	}
}

// GetTokenInfo returns token information for a given token identifier
func (tm *TokenMappingConfig) GetTokenInfo(tokenIdentifier string) (*TokenInfo, error) {
	if tokenIdentifier == "" {
		// Default to ETH if no token specified
		return tm.tokens["ETH"], nil
	}
	
	normalizedToken := strings.ToUpper(strings.TrimSpace(tokenIdentifier))
	
	if tokenInfo, exists := tm.tokens[normalizedToken]; exists {
		return tokenInfo, nil
	}
	
	return nil, fmt.Errorf("unsupported token identifier: %s", tokenIdentifier)
}

// GetChainForToken returns the target chain name for a given token identifier
func (tm *TokenMappingConfig) GetChainForToken(tokenIdentifier string) (string, error) {
	tokenInfo, err := tm.GetTokenInfo(tokenIdentifier)
	if err != nil {
		return "", err
	}
	
	return tokenInfo.ChainName, nil
}

// IsNativeToken checks if the given token is the native token for its chain
func (tm *TokenMappingConfig) IsNativeToken(tokenIdentifier string) (bool, error) {
	tokenInfo, err := tm.GetTokenInfo(tokenIdentifier)
	if err != nil {
		return false, err
	}
	
	return tokenInfo.IsNative, nil
}

// GetSupportedTokens returns a list of all supported token identifiers
func (tm *TokenMappingConfig) GetSupportedTokens() []string {
	tokens := make([]string, 0, len(tm.tokens))
	
	for tokenId := range tm.tokens {
		tokens = append(tokens, tokenId)
	}
	
	return tokens
}

// GetNativeTokensPerChain returns a map of chain names to their native token symbols
func (tm *TokenMappingConfig) GetNativeTokensPerChain() map[string]string {
	nativeTokens := make(map[string]string)
	
	for _, tokenInfo := range tm.tokens {
		if tokenInfo.IsNative {
			nativeTokens[tokenInfo.ChainName] = tokenInfo.Symbol
		}
	}
	
	return nativeTokens
}

// ValidateTokenForChain checks if a token is supported on a specific chain
func (tm *TokenMappingConfig) ValidateTokenForChain(tokenIdentifier, chainName string) error {
	tokenInfo, err := tm.GetTokenInfo(tokenIdentifier)
	if err != nil {
		return err
	}
	
	expectedChain := tokenInfo.ChainName
	normalizedChain := strings.ToUpper(strings.TrimSpace(chainName))
	normalizedExpected := strings.ToUpper(strings.TrimSpace(expectedChain))
	
	if normalizedChain != normalizedExpected {
		return fmt.Errorf("token %s is not supported on chain %s, expected chain: %s", 
			tokenIdentifier, chainName, expectedChain)
	}
	
	return nil
}

// AddCustomToken allows adding custom token mappings (for testing or extensions)
func (tm *TokenMappingConfig) AddCustomToken(symbol string, aliases []string, chainName string, isNative bool, decimals int, description string) {
	token := &TokenInfo{
		Symbol:      symbol,
		Aliases:     aliases,
		ChainName:   chainName,
		IsNative:    isNative,
		Decimals:    decimals,
		Description: description,
	}
	
	tm.registerToken(token)
}

// Global instance for easy access
var DefaultTokenMapping = NewTokenMappingConfig()