// SPDX-License-Identifier: Apache-2.0
package dex

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

// Uniswap V2 constants for Ethereum mainnet
const (
	// Uniswap V2 Router address on Ethereum
	UniswapV2RouterAddress = "0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D"
	
	// WETH address on Ethereum (for ETH-token swaps)
	WETHAddress = "0xC02aaA39b223FE65608aaC29B7b0Ed646Cf8F7D5"
	
	// Common token addresses for validation
	USDCAddress = "0xA0b86a33E6441e064c7d56Fb65dd0E8F8b764C4e"
	USDTAddress = "0xdAC17F958D2ee523a2206206994597C13D831ec7"
	DAIAddress  = "0x6B175474E89094C44Da98b954EedeAC495271d0F"
)

// UniswapV2 implements the IDEX interface for Uniswap V2 protocol
type UniswapV2 struct {
	chainName     string
	routerAddress string
	wethAddress   string
}

// NewUniswapV2 creates a new Uniswap V2 DEX instance
func NewUniswapV2(chainName string) *UniswapV2 {
	var routerAddr, wethAddr string
	
	switch strings.ToLower(chainName) {
	case "ethereum", "eth":
		routerAddr = UniswapV2RouterAddress
		wethAddr = WETHAddress
	default:
		// For now, only support Ethereum
		routerAddr = UniswapV2RouterAddress
		wethAddr = WETHAddress
	}
	
	return &UniswapV2{
		chainName:     chainName,
		routerAddress: routerAddr,
		wethAddress:   wethAddr,
	}
}

// GetProtocolName returns the protocol name
func (u *UniswapV2) GetProtocolName() string {
	return "uniswap_v2"
}

// GetChainName returns the blockchain this DEX operates on
func (u *UniswapV2) GetChainName() string {
	return u.chainName
}

// GetQuote returns a price quote for token swap
func (u *UniswapV2) GetQuote(ctx context.Context, params *SwapParams) (*SwapQuote, error) {
	if err := u.validateSwapParams(params); err != nil {
		return nil, fmt.Errorf("invalid swap parameters: %w", err)
	}
	
	// For MVP implementation, return mock quote
	// In production, this would query Uniswap V2 pairs and calculate actual amounts
	var amountIn, amountOut *big.Int
	
	if params.AmountIn != nil {
		// Exact input swap
		amountIn = params.AmountIn
		// Mock calculation: assume 1:2500 ratio for USDC (6 decimals) to some token (18 decimals)
		amountOut = new(big.Int).Mul(params.AmountIn, big.NewInt(2500000000000000)) // Convert 6 decimals to 18
	} else if params.AmountOut != nil {
		// Exact output swap  
		amountOut = params.AmountOut
		// Mock calculation: reverse of above
		amountIn = new(big.Int).Div(params.AmountOut, big.NewInt(2500000000000000))
	} else {
		return nil, errors.New("either AmountIn or AmountOut must be specified")
	}
	
	// Calculate minimum output amount with slippage
	slippageMultiplier := big.NewInt(int64((100.0 - params.SlippageTolerance) * 100))
	amountOutMin := new(big.Int).Mul(amountOut, slippageMultiplier)
	amountOutMin.Div(amountOutMin, big.NewInt(10000))
	
	// Build route (direct pair for simplicity)
	route := []string{params.TokenIn, params.TokenOut}
	
	return &SwapQuote{
		AmountIn:     amountIn,
		AmountOut:    amountOut,
		AmountOutMin: amountOutMin,
		PriceImpact:  0.15, // Mock 0.15% price impact
		Route:        route,
		GasEstimate:  150000, // Mock gas estimate
	}, nil
}

// SwapExactTokensForTokens performs exact input token swap
func (u *UniswapV2) SwapExactTokensForTokens(ctx context.Context, params *SwapParams) (*SwapResult, error) {
	if params.AmountIn == nil {
		return nil, errors.New("AmountIn must be specified for exact input swap")
	}
	
	// Get quote first
	quote, err := u.GetQuote(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get quote: %w", err)
	}
	
	// In a real implementation, this would:
	// 1. Check and handle token approvals
	// 2. Build the transaction data for Uniswap V2 Router
	// 3. Sign and send the transaction
	// 4. Wait for confirmation and return actual results
	
	// For MVP, return mock successful result
	mockTxHash := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	
	return &SwapResult{
		TransactionHash: mockTxHash,
		AmountIn:        quote.AmountIn,
		AmountOut:       quote.AmountOut,
		PriceImpact:     quote.PriceImpact,
		GasUsed:         140000, // Slightly less than estimate
		Route:           quote.Route,
	}, nil
}

// SwapTokensForExactTokens performs exact output token swap
func (u *UniswapV2) SwapTokensForExactTokens(ctx context.Context, params *SwapParams) (*SwapResult, error) {
	if params.AmountOut == nil {
		return nil, errors.New("AmountOut must be specified for exact output swap")
	}
	
	// Get quote first
	quote, err := u.GetQuote(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get quote: %w", err)
	}
	
	// For MVP, return mock successful result
	mockTxHash := "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	
	return &SwapResult{
		TransactionHash: mockTxHash,
		AmountIn:        quote.AmountIn,
		AmountOut:       quote.AmountOut,
		PriceImpact:     quote.PriceImpact,
		GasUsed:         140000,
		Route:           quote.Route,
	}, nil
}

// GetSupportedTokens returns commonly supported tokens on Uniswap V2
func (u *UniswapV2) GetSupportedTokens(ctx context.Context) ([]string, error) {
	// Return well-known token addresses for Ethereum
	supportedTokens := []string{
		WETHAddress, // WETH
		USDCAddress, // USDC  
		USDTAddress, // USDT
		DAIAddress,  // DAI
	}
	
	return supportedTokens, nil
}

// validateSwapParams validates the swap parameters
func (u *UniswapV2) validateSwapParams(params *SwapParams) error {
	if params == nil {
		return errors.New("swap parameters cannot be nil")
	}
	
	// Validate token addresses
	if !common.IsHexAddress(params.TokenIn) {
		return fmt.Errorf("invalid TokenIn address: %s", params.TokenIn)
	}
	
	if !common.IsHexAddress(params.TokenOut) {
		return fmt.Errorf("invalid TokenOut address: %s", params.TokenOut)
	}
	
	// Tokens must be different
	if strings.EqualFold(params.TokenIn, params.TokenOut) {
		return errors.New("TokenIn and TokenOut must be different")
	}
	
	// Validate recipient address
	if !common.IsHexAddress(params.Recipient) {
		return fmt.Errorf("invalid recipient address: %s", params.Recipient)
	}
	
	// Validate slippage tolerance (0.1% to 50%)
	if params.SlippageTolerance < 0.1 || params.SlippageTolerance > 50.0 {
		return fmt.Errorf("slippage tolerance must be between 0.1%% and 50%%, got %.2f%%", params.SlippageTolerance)
	}
	
	// Validate amounts
	if params.AmountIn == nil && params.AmountOut == nil {
		return errors.New("either AmountIn or AmountOut must be specified")
	}
	
	if params.AmountIn != nil && params.AmountOut != nil {
		return errors.New("only one of AmountIn or AmountOut should be specified")
	}
	
	if params.AmountIn != nil && params.AmountIn.Cmp(big.NewInt(0)) <= 0 {
		return errors.New("AmountIn must be positive")
	}
	
	if params.AmountOut != nil && params.AmountOut.Cmp(big.NewInt(0)) <= 0 {
		return errors.New("AmountOut must be positive")
	}
	
	return nil
}