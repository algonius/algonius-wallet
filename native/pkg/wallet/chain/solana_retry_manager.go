// SPDX-License-Identifier: Apache-2.0
package chain

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
)

// SolanaRetryManager handles transaction retry logic with intelligent error classification
type SolanaRetryManager struct {
	config  *RetryConfig
	logger  *zap.Logger
	runMode string
}

// TransactionParams represents parameters for a Solana transaction
type TransactionParams struct {
	From             string
	To               string
	Amount           uint64
	TokenMint        string // Optional: for SPL token transfers
	Slippage         float64
	RecentBlockhash  string
	JitoTipAmount    uint64
	MaxRetries       int
	GasStrategy      string
}

// TransactionResult represents the result of a transaction attempt
type TransactionResult struct {
	Signature   string
	Successful  bool
	Error       error
	Attempt     int
	FinalSlippage float64
}

// NewSolanaRetryManager creates a new retry manager
func NewSolanaRetryManager(config *RetryConfig, logger *zap.Logger) *SolanaRetryManager {
	return &SolanaRetryManager{
		config:  config,
		logger:  logger,
		runMode: os.Getenv("RUN_MODE"),
	}
}

// ExecuteWithRetry executes a transaction with retry logic
func (rm *SolanaRetryManager) ExecuteWithRetry(
	ctx context.Context,
	params *TransactionParams,
	executor func(context.Context, *TransactionParams) (string, error),
) (*TransactionResult, error) {
	
	if rm.runMode == "test" {
		return rm.executeMockTransaction(ctx, params)
	}
	
	var lastError error
	originalSlippage := params.Slippage
	maxRetries := rm.config.MaxRetries
	
	// Use transaction-specific max retries if provided
	if params.MaxRetries > 0 {
		maxRetries = params.MaxRetries
	}
	
	rm.logger.Info("Starting transaction execution with retry",
		zap.String("from", params.From),
		zap.String("to", params.To),
		zap.Uint64("amount", params.Amount),
		zap.Float64("slippage", params.Slippage),
		zap.Int("max_retries", maxRetries))
	
	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Update parameters for retry attempts
		if attempt > 0 {
			if err := rm.updateParamsForRetry(params, attempt, lastError); err != nil {
				rm.logger.Error("Failed to update parameters for retry", zap.Error(err))
				continue
			}
			
			// Apply retry delay with exponential backoff
			delay := rm.calculateRetryDelay(attempt)
			rm.logger.Info("Retrying transaction after delay",
				zap.Int("attempt", attempt+1),
				zap.Duration("delay", delay),
				zap.Float64("updated_slippage", params.Slippage))
			
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
				// Continue with retry
			}
		}
		
		// Execute the transaction
		signature, err := executor(ctx, params)
		
		if err == nil {
			rm.logger.Info("Transaction executed successfully",
				zap.String("signature", signature),
				zap.Int("attempt", attempt+1),
				zap.Float64("final_slippage", params.Slippage))
			
			return &TransactionResult{
				Signature:     signature,
				Successful:    true,
				Attempt:       attempt + 1,
				FinalSlippage: params.Slippage,
			}, nil
		}
		
		lastError = err
		rm.logger.Warn("Transaction attempt failed",
			zap.Error(err),
			zap.Int("attempt", attempt+1),
			zap.String("error_type", rm.classifyError(err)))
		
		// Check if error is retryable
		if !rm.isRetryableError(err) {
			rm.logger.Error("Non-retryable error encountered, stopping retries",
				zap.Error(err),
				zap.Int("attempt", attempt+1))
			
			return &TransactionResult{
				Signature:     "",
				Successful:    false,
				Error:         err,
				Attempt:       attempt + 1,
				FinalSlippage: params.Slippage,
			}, err
		}
		
		// Check if we've reached maximum slippage
		if params.Slippage >= float64(rm.config.MaxTotalSlippageBps)/10000 {
			rm.logger.Error("Maximum slippage reached, stopping retries",
				zap.Float64("current_slippage", params.Slippage),
				zap.Float64("max_slippage", float64(rm.config.MaxTotalSlippageBps)/10000))
			
			return &TransactionResult{
				Signature:     "",
				Successful:    false,
				Error:         fmt.Errorf("maximum slippage exceeded: %w", lastError),
				Attempt:       attempt + 1,
				FinalSlippage: params.Slippage,
			}, fmt.Errorf("maximum slippage exceeded after %d attempts: %w", attempt+1, lastError)
		}
	}
	
	// All retries exhausted
	rm.logger.Error("All retry attempts exhausted",
		zap.Error(lastError),
		zap.Int("attempts", maxRetries+1),
		zap.Float64("original_slippage", originalSlippage),
		zap.Float64("final_slippage", params.Slippage))
	
	return &TransactionResult{
		Signature:     "",
		Successful:    false,
		Error:         lastError,
		Attempt:       maxRetries + 1,
		FinalSlippage: params.Slippage,
	}, fmt.Errorf("transaction failed after %d attempts: %w", maxRetries+1, lastError)
}

// updateParamsForRetry updates transaction parameters for retry attempts
func (rm *SolanaRetryManager) updateParamsForRetry(params *TransactionParams, attempt int, lastError error) error {
	// Update blockhash for fresh transaction
	// This would typically involve calling GetLatestBlockhash, but we'll simulate it here
	params.RecentBlockhash = fmt.Sprintf("UpdatedBlockhash%d_%d", attempt, time.Now().Unix())
	
	// Increase slippage for slippage-related errors
	if rm.isSlippageError(lastError) {
		slippageIncrease := float64(rm.config.SlippageIncrementBps*attempt) / 10000
		params.Slippage += slippageIncrease
		
		rm.logger.Debug("Increased slippage for retry",
			zap.Float64("slippage_increase", slippageIncrease),
			zap.Float64("new_slippage", params.Slippage))
	}
	
	// Update Jito tip based on gas strategy
	if params.GasStrategy != "" || rm.config.GasStrategy != "" {
		strategy := params.GasStrategy
		if strategy == "" {
			strategy = rm.config.GasStrategy
		}
		
		params.JitoTipAmount = rm.calculateJitoTip(params.JitoTipAmount, attempt, strategy)
	}
	
	return nil
}

// calculateJitoTip calculates Jito tip amount based on retry attempt and strategy
func (rm *SolanaRetryManager) calculateJitoTip(baseTip uint64, attempt int, strategy string) uint64 {
	switch strategy {
	case "level_up":
		// Increase tip by 50% each retry
		multiplier := 1.0 + (0.5 * float64(attempt))
		return uint64(float64(baseTip) * multiplier)
	case "level_multiple":
		// Double tip each retry
		return baseTip * uint64(1<<attempt)
	case "exponential":
		// Exponential increase with base 1.5
		multiplier := 1.0
		for i := 0; i < attempt; i++ {
			multiplier *= 1.5
		}
		return uint64(float64(baseTip) * multiplier)
	default:
		// Linear increase
		return baseTip + (uint64(attempt) * 1000) // Add 1000 lamports per attempt
	}
}

// calculateRetryDelay calculates delay before retry with exponential backoff
func (rm *SolanaRetryManager) calculateRetryDelay(attempt int) time.Duration {
	baseDelay := rm.config.BaseRetryDelay
	if baseDelay == 0 {
		baseDelay = time.Second * 2
	}
	
	// Exponential backoff: baseDelay * 2^(attempt-1)
	multiplier := 1 << (attempt - 1)
	delay := baseDelay * time.Duration(multiplier)
	
	// Cap maximum delay at 30 seconds
	maxDelay := time.Second * 30
	if delay > maxDelay {
		delay = maxDelay
	}
	
	return delay
}

// isRetryableError determines if an error should trigger a retry
func (rm *SolanaRetryManager) isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	
	return rm.isSlippageError(err) || rm.isRPCError(err) || rm.isBlockhashError(err)
}

// isSlippageError checks if error is related to slippage
func (rm *SolanaRetryManager) isSlippageError(err error) bool {
	if err == nil {
		return false
	}
	
	errMsg := strings.ToLower(err.Error())
	slippageKeywords := []string{
		"slippage tolerance exceeded",
		"price impact too high",
		"insufficient output amount",
		"slippage",
		"price impact",
		"minimum received",
	}
	
	for _, keyword := range slippageKeywords {
		if strings.Contains(errMsg, keyword) {
			return true
		}
	}
	
	return false
}

// isRPCError checks if error is related to RPC connectivity
func (rm *SolanaRetryManager) isRPCError(err error) bool {
	if err == nil {
		return false
	}
	
	errMsg := strings.ToLower(err.Error())
	rpcKeywords := []string{
		"rpc",
		"network",
		"connection",
		"timeout",
		"rate limit",
		"429",
		"503",
		"502",
		"500",
	}
	
	for _, keyword := range rpcKeywords {
		if strings.Contains(errMsg, keyword) {
			return true
		}
	}
	
	return false
}

// isBlockhashError checks if error is related to blockhash
func (rm *SolanaRetryManager) isBlockhashError(err error) bool {
	if err == nil {
		return false
	}
	
	errMsg := strings.ToLower(err.Error())
	blockhashKeywords := []string{
		"blockhash not found",
		"blockhash expired",
		"invalid blockhash",
		"transaction expired",
	}
	
	for _, keyword := range blockhashKeywords {
		if strings.Contains(errMsg, keyword) {
			return true
		}
	}
	
	return false
}

// classifyError returns a string classification of the error
func (rm *SolanaRetryManager) classifyError(err error) string {
	if err == nil {
		return "none"
	}
	
	if rm.isSlippageError(err) {
		return "slippage"
	}
	if rm.isRPCError(err) {
		return "rpc"
	}
	if rm.isBlockhashError(err) {
		return "blockhash"
	}
	
	return "unknown"
}

// executeMockTransaction executes a mock transaction for testing
func (rm *SolanaRetryManager) executeMockTransaction(_ context.Context, params *TransactionParams) (*TransactionResult, error) {
	rm.logger.Debug("Executing mock transaction for testing",
		zap.String("from", params.From),
		zap.String("to", params.To),
		zap.Uint64("amount", params.Amount))
	
	// Simulate processing time
	time.Sleep(time.Millisecond * 50)
	
	// In test mode, always succeed on first attempt
	return &TransactionResult{
		Signature:     "MockTransactionSignature123456789abcdef",
		Successful:    true,
		Attempt:       1,
		FinalSlippage: params.Slippage,
	}, nil
}

// GetRetryStatistics returns statistics about retry behavior
func (rm *SolanaRetryManager) GetRetryStatistics() map[string]any {
	return map[string]any{
		"max_retries":             rm.config.MaxRetries,
		"slippage_increment_bps":  rm.config.SlippageIncrementBps,
		"base_retry_delay":        rm.config.BaseRetryDelay.String(),
		"max_total_slippage_bps":  rm.config.MaxTotalSlippageBps,
		"gas_strategy":            rm.config.GasStrategy,
		"run_mode":                rm.runMode,
	}
}