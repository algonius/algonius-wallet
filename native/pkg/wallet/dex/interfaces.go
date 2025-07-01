// SPDX-License-Identifier: Apache-2.0
package dex

import (
	"context"
	"math/big"
)

// SwapParams contains parameters for token swap operations
type SwapParams struct {
	// Input token parameters
	TokenIn     string   // Input token contract address
	TokenOut    string   // Output token contract address
	AmountIn    *big.Int // Input amount (for exact input swaps)
	AmountOut   *big.Int // Output amount (for exact output swaps)
	
	// Swap configuration
	SlippageTolerance float64 // Maximum slippage tolerance (0.5 = 0.5%)
	Deadline          uint64  // Transaction deadline (unix timestamp)
	Recipient         string  // Recipient address
	
	// Transaction details
	From       string // Sender address
	PrivateKey string // Private key for signing
}

// SwapQuote represents a price quote for a token swap
type SwapQuote struct {
	AmountIn     *big.Int // Input amount
	AmountOut    *big.Int // Expected output amount
	AmountOutMin *big.Int // Minimum output amount after slippage
	PriceImpact  float64  // Price impact percentage
	Route        []string // Token addresses in swap route
	GasEstimate  uint64   // Estimated gas for the swap
}

// SwapResult contains the result of a successful swap transaction
type SwapResult struct {
	TransactionHash string   // Transaction hash
	AmountIn        *big.Int // Actual input amount
	AmountOut       *big.Int // Actual output amount
	PriceImpact     float64  // Actual price impact
	GasUsed         uint64   // Gas used by transaction
	Route           []string // Actual swap route used
}

// IDEX defines the interface for decentralized exchange integrations
type IDEX interface {
	// GetQuote returns a quote for swapping tokens
	GetQuote(ctx context.Context, params *SwapParams) (*SwapQuote, error)
	
	// SwapExactTokensForTokens performs exact input swap
	SwapExactTokensForTokens(ctx context.Context, params *SwapParams) (*SwapResult, error)
	
	// SwapTokensForExactTokens performs exact output swap
	SwapTokensForExactTokens(ctx context.Context, params *SwapParams) (*SwapResult, error)
	
	// GetSupportedTokens returns list of supported token addresses
	GetSupportedTokens(ctx context.Context) ([]string, error)
	
	// GetProtocolName returns the name of the DEX protocol
	GetProtocolName() string
	
	// GetChainName returns the blockchain this DEX operates on
	GetChainName() string
}

// ITokenValidator provides token validation functionality
type ITokenValidator interface {
	// ValidateTokenAddress checks if a token address is valid
	ValidateTokenAddress(ctx context.Context, tokenAddress string) error
	
	// GetTokenDecimals returns the decimal places for a token
	GetTokenDecimals(ctx context.Context, tokenAddress string) (uint8, error)
	
	// GetTokenSymbol returns the symbol for a token
	GetTokenSymbol(ctx context.Context, tokenAddress string) (string, error)
}

// IDEXFactory creates DEX instances for different protocols and chains
type IDEXFactory interface {
	// CreateDEX creates a DEX instance for the specified protocol and chain
	CreateDEX(protocol string, chain string) (IDEX, error)
	
	// GetSupportedProtocols returns supported DEX protocols
	GetSupportedProtocols() []string
	
	// GetSupportedChains returns supported blockchain networks
	GetSupportedChains(protocol string) []string
}