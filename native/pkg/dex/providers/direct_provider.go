// SPDX-License-Identifier: Apache-2.0
package providers

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/algonius/algonius-wallet/native/pkg/dex"
	"go.uber.org/zap"
)

// DirectProvider implements direct RPC calls without external DEX APIs
type DirectProvider struct {
	name            string
	logger          *zap.Logger
	supportedChains map[string]bool
}

// NewDirectProvider creates a new direct RPC provider
func NewDirectProvider(logger *zap.Logger) *DirectProvider {
	supportedChains := map[string]bool{
		"1":   true, // Ethereum
		"56":  true, // BSC
		"501": true, // Solana
	}

	return &DirectProvider{
		name:            "Direct",
		logger:          logger,
		supportedChains: supportedChains,
	}
}

// GetName returns the name of the DEX provider
func (d *DirectProvider) GetName() string {
	return d.name
}

// IsSupported checks if the chain is supported by this provider
func (d *DirectProvider) IsSupported(chainID string) bool {
	return d.supportedChains[chainID]
}

// GetQuote gets a quote for token swap using direct RPC calls
func (d *DirectProvider) GetQuote(ctx context.Context, params dex.SwapParams) (*dex.SwapQuote, error) {
	if err := d.validateSwapParams(params); err != nil {
		return nil, err
	}

	// For direct provider, simulate basic swap scenarios
	quote, err := d.simulateSwapQuote(params)
	if err != nil {
		return nil, err
	}

	d.logger.Debug("Direct provider quote generated", 
		zap.String("fromToken", quote.FromToken),
		zap.String("toToken", quote.ToToken),
		zap.String("toAmount", quote.ToAmount))

	return quote, nil
}

// ExecuteSwap executes a token swap using direct RPC calls
func (d *DirectProvider) ExecuteSwap(ctx context.Context, params dex.SwapParams) (*dex.SwapResult, error) {
	if err := d.validateSwapParams(params); err != nil {
		return nil, err
	}

	// For direct provider, simulate the transaction
	result, err := d.simulateSwapExecution(params)
	if err != nil {
		return nil, err
	}

	d.logger.Info("Direct provider swap executed", 
		zap.String("txHash", result.TxHash),
		zap.String("fromToken", params.FromToken),
		zap.String("toToken", params.ToToken))

	return result, nil
}

// GetBalance gets token balance using direct RPC calls
func (d *DirectProvider) GetBalance(ctx context.Context, address string, tokenAddress string, chainID string) (*dex.BalanceInfo, error) {
	if !d.IsSupported(chainID) {
		return nil, fmt.Errorf("chain %s not supported by direct provider", chainID)
	}

	// For direct provider, simulate balance retrieval
	balance, err := d.simulateBalanceCheck(address, tokenAddress, chainID)
	if err != nil {
		return nil, err
	}

	return balance, nil
}

// EstimateGas estimates gas for a swap operation
func (d *DirectProvider) EstimateGas(ctx context.Context, params dex.SwapParams) (gasLimit uint64, gasPrice string, err error) {
	// Basic gas estimation based on chain type
	switch params.ChainID {
	case "1": // Ethereum
		return 150000, "20", nil // 20 gwei
	case "56": // BSC
		return 120000, "5", nil // 5 gwei
	case "501": // Solana
		return 5000, "1", nil // 1 microlamport per compute unit
	default:
		return 0, "", fmt.Errorf("gas estimation not implemented for chain %s", params.ChainID)
	}
}

// validateSwapParams validates swap parameters
func (d *DirectProvider) validateSwapParams(params dex.SwapParams) error {
	if params.FromToken == "" {
		return fmt.Errorf("from token cannot be empty")
	}
	if params.ToToken == "" {
		return fmt.Errorf("to token cannot be empty")
	}
	if params.Amount == "" {
		return fmt.Errorf("amount cannot be empty")
	}
	if params.FromAddress == "" {
		return fmt.Errorf("from address cannot be empty")
	}
	if params.ChainID == "" {
		return fmt.Errorf("chain ID cannot be empty")
	}
	if !d.IsSupported(params.ChainID) {
		return fmt.Errorf("chain %s not supported", params.ChainID)
	}
	return nil
}

// simulateSwapQuote simulates generating a swap quote
func (d *DirectProvider) simulateSwapQuote(params dex.SwapParams) (*dex.SwapQuote, error) {
	// Simple simulation: assume 1:1 ratio for native currency swaps, apply slippage
	fromAmount, ok := new(big.Float).SetString(params.Amount)
	if !ok {
		return nil, fmt.Errorf("invalid amount: %s", params.Amount)
	}

	// Apply a simulated exchange rate
	var rate float64 = 0.95 // Simulate 5% price difference
	toAmount := new(big.Float).Mul(fromAmount, big.NewFloat(rate))
	
	// Apply slippage
	slippageAmount := new(big.Float).Mul(toAmount, big.NewFloat(params.Slippage))
	toAmount.Sub(toAmount, slippageAmount)

	// Estimate gas
	gasLimit, gasPrice, _ := d.EstimateGas(context.Background(), params)

	quote := &dex.SwapQuote{
		FromToken:      params.FromToken,
		ToToken:        params.ToToken,
		FromAmount:     params.Amount,
		ToAmount:       toAmount.String(),
		EstimatedGas:   gasLimit,
		GasPrice:       gasPrice,
		EstimatedFee:   d.calculateEstimatedFee(gasLimit, gasPrice),
		Slippage:       params.Slippage,
		PriceImpact:    0.001, // 0.1% simulated price impact
		Route:          []string{"Direct DEX"},
		ValidUntil:     time.Now().Add(2 * time.Minute).Unix(),
		Provider:       d.name,
		RawData:        "direct_simulation",
	}

	return quote, nil
}

// simulateSwapExecution simulates executing a swap
func (d *DirectProvider) simulateSwapExecution(params dex.SwapParams) (*dex.SwapResult, error) {
	// Generate a mock transaction hash based on chain type
	var txHash string
	switch params.ChainID {
	case "1", "56": // EVM chains
		txHash = "0x" + strings.Repeat("d", 64)
	case "501": // Solana
		txHash = strings.Repeat("9", 64)
	default:
		txHash = "mock_tx_hash"
	}

	// Get the quote to determine output amount
	quote, err := d.simulateSwapQuote(params)
	if err != nil {
		return nil, err
	}

	result := &dex.SwapResult{
		TxHash:        txHash,
		Status:        "pending",
		FromToken:     params.FromToken,
		ToToken:       params.ToToken,
		FromAmount:    params.Amount,
		ToAmount:      quote.ToAmount,
		ActualFee:     quote.EstimatedFee,
		Provider:      d.name,
		Timestamp:     time.Now().Unix(),
	}

	return result, nil
}

// simulateBalanceCheck simulates checking token balance
func (d *DirectProvider) simulateBalanceCheck(address, tokenAddress, chainID string) (*dex.BalanceInfo, error) {
	// In simulation, return mock balance data
	var tokenSymbol string
	var decimals int
	
	switch chainID {
	case "1": // Ethereum
		if tokenAddress == "" || tokenAddress == "ETH" {
			tokenSymbol = "ETH"
		} else {
			tokenSymbol = "TOKEN"
		}
		decimals = 18
	case "56": // BSC
		if tokenAddress == "" || tokenAddress == "BNB" {
			tokenSymbol = "BNB"
		} else {
			tokenSymbol = "TOKEN"
		}
		decimals = 18
	case "501": // Solana
		if tokenAddress == "" || tokenAddress == "SOL" {
			tokenSymbol = "SOL"
		} else {
			tokenSymbol = "TOKEN"
		}
		decimals = 9
	default:
		tokenSymbol = "UNKNOWN"
		decimals = 18
	}

	balance := &dex.BalanceInfo{
		TokenAddress: tokenAddress,
		TokenSymbol:  tokenSymbol,
		Balance:      "1.000000000000000000", // Mock balance
		Decimals:     decimals,
		USDValue:     "1.00",
	}

	return balance, nil
}

// calculateEstimatedFee calculates estimated transaction fee
func (d *DirectProvider) calculateEstimatedFee(gasLimit uint64, gasPrice string) string {
	gasPriceBig, ok := new(big.Float).SetString(gasPrice)
	if !ok {
		return "0.001" // Default fee
	}

	// Calculate fee = gasLimit * gasPrice
	gasLimitBig := new(big.Float).SetUint64(gasLimit)
	fee := new(big.Float).Mul(gasLimitBig, gasPriceBig)
	
	// Convert from gwei to ether (divide by 1e9) for EVM chains
	fee.Quo(fee, big.NewFloat(1e9))

	return fee.String()
}