// Package simulation provides transaction simulation capabilities for the Algonius Wallet.
package simulation

import (
	"context"
	"fmt"
	"math/big"

	"github.com/algonius/algonius-wallet/native/pkg/wallet/chain"
)

// SimulationResult represents the result of a transaction simulation
type SimulationResult struct {
	Success      bool     `json:"success"`
	GasUsed      uint64   `json:"gas_used"`
	GasPrice     string   `json:"gas_price"`
	TotalCost    string   `json:"total_cost"`
	BalanceChange string  `json:"balance_change"`
	Warnings     []string `json:"warnings"`
	Errors       []string `json:"errors"`
}

// TransactionSimulator handles transaction simulations
type TransactionSimulator struct {
	chainFactory *chain.ChainFactory
}

// NewTransactionSimulator creates a new TransactionSimulator
func NewTransactionSimulator(chainFactory *chain.ChainFactory) *TransactionSimulator {
	return &TransactionSimulator{
		chainFactory: chainFactory,
	}
}

// SimulateTransaction simulates a transaction without executing it
func (s *TransactionSimulator) SimulateTransaction(ctx context.Context, chainName, from, to, amount, token string) (*SimulationResult, error) {
	// Get chain implementation
	chainImpl, err := s.chainFactory.GetChain(chainName)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain implementation: %w", err)
	}

	// Note: Skipping address validation for now as it's not implemented in IChain interface
	// In a real implementation, you would validate addresses using chain-specific methods

	// Parse amount
	amountValue, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return nil, fmt.Errorf("invalid amount: %s", amount)
	}

	// Check balance
	balance, err := chainImpl.GetBalance(ctx, from, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	balanceValue, ok := new(big.Int).SetString(balance, 10)
	if !ok {
		return nil, fmt.Errorf("invalid balance format: %s", balance)
	}

	// Check if sufficient funds
	if balanceValue.Cmp(amountValue) < 0 {
		return &SimulationResult{
			Success: false,
			Errors:  []string{"Insufficient funds"},
		}, nil
	}

	// Estimate gas
	gasLimit, gasPrice, err := chainImpl.EstimateGas(ctx, from, to, amount, token)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate gas: %w", err)
	}

	// Calculate total cost
	gasPriceValue, ok := new(big.Int).SetString(gasPrice, 10)
	if !ok {
		return nil, fmt.Errorf("invalid gas price format: %s", gasPrice)
	}

	gasCost := new(big.Int).Mul(big.NewInt(int64(gasLimit)), gasPriceValue)
	totalCost := new(big.Int).Add(amountValue, gasCost)

	// Check if sufficient funds for total cost
	if balanceValue.Cmp(totalCost) < 0 {
		return &SimulationResult{
			Success:   false,
			GasUsed:   gasLimit,
			GasPrice:  gasPrice,
			TotalCost: totalCost.String(),
			Errors:    []string{"Insufficient funds for transaction + gas fees"},
		}, nil
	}

	// Check for potential warnings
	var warnings []string
	if gasPriceValue.Cmp(big.NewInt(100)) > 0 {
		warnings = append(warnings, "High gas price detected")
	}

	// Calculate balance change
	balanceChange := new(big.Int).Neg(totalCost)

	return &SimulationResult{
		Success:      true,
		GasUsed:      gasLimit,
		GasPrice:     gasPrice,
		TotalCost:    totalCost.String(),
		BalanceChange: balanceChange.String(),
		Warnings:     warnings,
		Errors:       []string{},
	}, nil
}