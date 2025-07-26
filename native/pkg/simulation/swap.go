// Package simulation provides swap simulation capabilities for the Algonius Wallet.
package simulation

import (
	"context"
	"fmt"
	"math/big"

	"github.com/algonius/algonius-wallet/native/pkg/wallet/chain"
	"github.com/algonius/algonius-wallet/native/pkg/wallet/dex"
)

// SwapSimulationResult represents the result of a swap simulation
type SwapSimulationResult struct {
	Success       bool     `json:"success"`
	AmountIn      string   `json:"amount_in"`
	AmountOut     string   `json:"amount_out"`
	PriceImpact   float64  `json:"price_impact"`
	GasUsed       uint64   `json:"gas_used"`
	GasPrice      string   `json:"gas_price"`
	TotalCost     string   `json:"total_cost"`
	BalanceChange string   `json:"balance_change"`
	Route         []string `json:"route"`
	Warnings      []string `json:"warnings"`
	Errors        []string `json:"errors"`
}

// SwapSimulator handles swap simulations
type SwapSimulator struct {
	chainFactory *chain.ChainFactory
	dexFactory   dex.IDEXFactory
}

// NewSwapSimulator creates a new SwapSimulator
func NewSwapSimulator(chainFactory *chain.ChainFactory, dexFactory dex.IDEXFactory) *SwapSimulator {
	return &SwapSimulator{
		chainFactory: chainFactory,
		dexFactory:   dexFactory,
	}
}

// SimulateSwap simulates a token swap without executing it
func (s *SwapSimulator) SimulateSwap(ctx context.Context, chainName, tokenIn, tokenOut, amountIn, amountOut, from, dexProtocol string, slippageTolerance float64) (*SwapSimulationResult, error) {
	// Get chain implementation
	chainImpl, err := s.chainFactory.GetChain(chainName)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain implementation: %w", err)
	}

	// Validate addresses
	if !chainImpl.ValidateAddress(from) {
		return nil, fmt.Errorf("invalid sender address: %s", from)
	}

	// Validate tokens
	if tokenIn != "" && !chainImpl.ValidateTokenAddress(tokenIn) {
		return nil, fmt.Errorf("invalid input token address: %s", tokenIn)
	}

	if tokenOut != "" && !chainImpl.ValidateTokenAddress(tokenOut) {
		return nil, fmt.Errorf("invalid output token address: %s", tokenOut)
	}

	// Create DEX instance
	dexInstance, err := s.dexFactory.CreateDEX(dexProtocol, chainName)
	if err != nil {
		return nil, fmt.Errorf("failed to create DEX instance: %w", err)
	}

	// Parse amounts
	var amountInValue, amountOutValue *big.Int
	if amountIn != "" {
		amountInValue, ok := new(big.Int).SetString(amountIn, 10)
		if !ok {
			return nil, fmt.Errorf("invalid amount_in: %s", amountIn)
		}
	}

	if amountOut != "" {
		amountOutValue, ok := new(big.Int).SetString(amountOut, 10)
		if !ok {
			return nil, fmt.Errorf("invalid amount_out: %s", amountOut)
		}
	}

	// Check balance for token in
	balance, err := chainImpl.GetBalance(ctx, from, tokenIn)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	balanceValue, ok := new(big.Int).SetString(balance, 10)
	if !ok {
		return nil, fmt.Errorf("invalid balance format: %s", balance)
	}

	// For exact input swaps, check if sufficient funds
	if amountInValue != nil && balanceValue.Cmp(amountInValue) < 0 {
		return &SwapSimulationResult{
			Success: false,
			Errors:  []string{"Insufficient funds for input token"},
		}, nil
	}

	// Create swap parameters
	swapParams := &dex.SwapParams{
		TokenIn:           tokenIn,
		TokenOut:          tokenOut,
		AmountIn:          amountInValue,
		AmountOut:         amountOutValue,
		SlippageTolerance: slippageTolerance,
		Recipient:         from,
		From:              from,
		// Note: We don't need a real private key for simulation
		PrivateKey: "0x0000000000000000000000000000000000000000000000000000000000000001",
	}

	// Get quote for the swap
	quote, err := dexInstance.GetQuote(ctx, swapParams)
	if err != nil {
		return nil, fmt.Errorf("failed to get swap quote: %w", err)
	}

	// Estimate gas for the swap
	gasLimit, gasPrice, err := dexInstance.EstimateGas(ctx, swapParams)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate gas: %w", err)
	}

	// Calculate total cost
	gasPriceValue, ok := new(big.Int).SetString(gasPrice, 10)
	if !ok {
		return nil, fmt.Errorf("invalid gas price format: %s", gasPrice)
	}

	gasCost := new(big.Int).Mul(big.NewInt(int64(gasLimit)), gasPriceValue)
	
	// For exact input swaps, we know the amount in
	var totalCost *big.Int
	if amountInValue != nil {
		totalCost = new(big.Int).Add(amountInValue, gasCost)
	} else {
		// For exact output swaps, we know the amount out but not the exact amount in
		// This is a simplified calculation for simulation purposes
		totalCost = gasCost
	}

	// Check if sufficient funds for total cost
	if balanceValue.Cmp(totalCost) < 0 {
		return &SwapSimulationResult{
			Success:   false,
			GasUsed:   gasLimit,
			GasPrice:  gasPrice,
			TotalCost: totalCost.String(),
			Errors:    []string{"Insufficient funds for swap + gas fees"},
		}, nil
	}

	// Check for potential warnings
	var warnings []string
	if gasPriceValue.Cmp(big.NewInt(100)) > 0 {
		warnings = append(warnings, "High gas price detected")
	}

	if quote.PriceImpact > 5.0 {
		warnings = append(warnings, fmt.Sprintf("High price impact: %.2f%%", quote.PriceImpact))
	}

	// Calculate balance change (negative since we're spending)
	balanceChange := new(big.Int).Neg(totalCost)

	return &SwapSimulationResult{
		Success:       true,
		AmountIn:      quote.AmountIn.String(),
		AmountOut:     quote.AmountOut.String(),
		PriceImpact:   quote.PriceImpact,
		GasUsed:       gasLimit,
		GasPrice:      gasPrice,
		TotalCost:     totalCost.String(),
		BalanceChange: balanceChange.String(),
		Route:         quote.Route,
		Warnings:      warnings,
		Errors:        []string{},
	}, nil
}