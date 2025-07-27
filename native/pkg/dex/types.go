// SPDX-License-Identifier: Apache-2.0
package dex

import (
	"context"
)

// SwapParams contains parameters for token swap operations
type SwapParams struct {
	FromToken    string  `json:"from_token"`    // Token address or symbol to swap from
	ToToken      string  `json:"to_token"`      // Token address or symbol to swap to
	Amount       string  `json:"amount"`        // Amount to swap (in token units)
	Slippage     float64 `json:"slippage"`      // Maximum acceptable slippage (0.01 = 1%)
	FromAddress  string  `json:"from_address"`  // Sender address
	ToAddress    string  `json:"to_address"`    // Recipient address (can be same as from)
	ChainID      string  `json:"chain_id"`      // Blockchain chain ID
	PrivateKey   string  `json:"private_key"`   // Private key for signing (handled securely)
}

// SwapQuote contains quote information for a token swap
type SwapQuote struct {
	FromToken      string     `json:"from_token"`
	ToToken        string     `json:"to_token"`
	FromAmount     string     `json:"from_amount"`
	ToAmount       string     `json:"to_amount"`
	EstimatedGas   uint64     `json:"estimated_gas"`
	GasPrice       string     `json:"gas_price"`
	EstimatedFee   string     `json:"estimated_fee"`
	Slippage       float64    `json:"slippage"`
	PriceImpact    float64    `json:"price_impact"`
	Route          []string   `json:"route"`          // DEX route information
	ValidUntil     int64      `json:"valid_until"`    // Quote expiry timestamp
	Provider       string     `json:"provider"`       // DEX provider name
	RawData        string     `json:"raw_data"`       // Provider-specific raw response
}

// SwapResult contains the result of a completed swap
type SwapResult struct {
	TxHash        string `json:"tx_hash"`
	Status        string `json:"status"`         // "pending", "confirmed", "failed"
	FromToken     string `json:"from_token"`
	ToToken       string `json:"to_token"`
	FromAmount    string `json:"from_amount"`
	ToAmount      string `json:"to_amount"`
	ActualFee     string `json:"actual_fee"`
	Provider      string `json:"provider"`
	Timestamp     int64  `json:"timestamp"`
}

// BalanceInfo contains token balance information
type BalanceInfo struct {
	TokenAddress string `json:"token_address"`
	TokenSymbol  string `json:"token_symbol"`
	Balance      string `json:"balance"`
	Decimals     int    `json:"decimals"`
	USDValue     string `json:"usd_value,omitempty"`
}

// IDEXProvider defines the interface for DEX providers
type IDEXProvider interface {
	// GetName returns the name of the DEX provider
	GetName() string
	
	// IsSupported checks if the chain is supported by this provider
	IsSupported(chainID string) bool
	
	// GetQuote gets a quote for token swap
	GetQuote(ctx context.Context, params SwapParams) (*SwapQuote, error)
	
	// ExecuteSwap executes a token swap
	ExecuteSwap(ctx context.Context, params SwapParams) (*SwapResult, error)
	
	// GetBalance gets token balance for an address
	GetBalance(ctx context.Context, address string, tokenAddress string, chainID string) (*BalanceInfo, error)
	
	// EstimateGas estimates gas for a swap operation
	EstimateGas(ctx context.Context, params SwapParams) (gasLimit uint64, gasPrice string, err error)
}

// IDEXAggregator defines the interface for DEX aggregation
type IDEXAggregator interface {
	// RegisterProvider registers a new DEX provider
	RegisterProvider(provider IDEXProvider) error
	
	// GetBestQuote gets the best quote from all available providers
	GetBestQuote(ctx context.Context, params SwapParams) (*SwapQuote, error)
	
	// ExecuteSwapWithProvider executes swap using a specific provider
	ExecuteSwapWithProvider(ctx context.Context, providerName string, params SwapParams) (*SwapResult, error)
	
	// GetSupportedProviders returns list of providers supporting the chain
	GetSupportedProviders(chainID string) []string
	
	// GetProviderByName returns a specific provider by name
	GetProviderByName(name string) (IDEXProvider, error)
}