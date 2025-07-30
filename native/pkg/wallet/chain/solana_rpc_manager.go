// SPDX-License-Identifier: Apache-2.0
package chain

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// SolanaRPCManager manages multiple RPC endpoints with automatic failover
type SolanaRPCManager struct {
	httpClient *http.Client
	endpoints  []string
	currentIdx atomic.Uint32
	logger     *zap.Logger
	runMode    string
}

// RPCRequest represents a JSON-RPC request
type RPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

// RPCResponse represents a JSON-RPC response
type RPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

// RPCError represents a JSON-RPC error
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// BlockhashResult represents latest blockhash response
type BlockhashResult struct {
	Context struct {
		Slot uint64 `json:"slot"`
	} `json:"context"`
	Value struct {
		Blockhash            string `json:"blockhash"`
		LastValidBlockHeight uint64 `json:"lastValidBlockHeight"`
	} `json:"value"`
}

// BalanceResult represents balance response
type BalanceResult struct {
	Context struct {
		Slot uint64 `json:"slot"`
	} `json:"context"`
	Value uint64 `json:"value"`
}

// SignatureStatusResult represents signature status response
type SignatureStatusResult struct {
	Context struct {
		Slot uint64 `json:"slot"`
	} `json:"context"`
	Value []struct {
		Slot               uint64  `json:"slot"`
		Confirmations      *uint64 `json:"confirmations"`
		ConfirmationStatus string  `json:"confirmationStatus"`
		Err                *string `json:"err"`
	} `json:"value"`
}

// NewSolanaRPCManager creates a new RPC manager with failover support
func NewSolanaRPCManager(endpoints []string, logger *zap.Logger) (*SolanaRPCManager, error) {
	if len(endpoints) == 0 {
		return nil, fmt.Errorf("at least one RPC endpoint is required")
	}
	
	manager := &SolanaRPCManager{
		httpClient: &http.Client{
			Timeout: time.Second * 30,
		},
		endpoints: endpoints,
		logger:    logger,
		runMode:   os.Getenv("RUN_MODE"),
	}
	
	return manager, nil
}

// getCurrentEndpoint returns the current active RPC endpoint
func (rm *SolanaRPCManager) getCurrentEndpoint() string {
	idx := rm.currentIdx.Load() % uint32(len(rm.endpoints))
	return rm.endpoints[idx]
}

// switchToNext switches to the next available RPC endpoint
func (rm *SolanaRPCManager) switchToNext() {
	old := rm.currentIdx.Load()
	new := (old + 1) % uint32(len(rm.endpoints))
	
	if rm.currentIdx.CompareAndSwap(old, new) {
		rm.logger.Warn("Switched to backup RPC endpoint", 
			zap.String("from", rm.endpoints[old]),
			zap.String("to", rm.endpoints[new]))
	}
}

// callRPC makes a JSON-RPC call to the Solana network
func (rm *SolanaRPCManager) callRPC(ctx context.Context, method string, params interface{}, result interface{}) error {
	if rm.runMode == "test" {
		// In test mode, return mock responses
		return rm.executeMockOperation(method, result)
	}
	
	var lastErr error
	startIdx := rm.currentIdx.Load()
	
	// Try all endpoints
	for i := 0; i < len(rm.endpoints); i++ {
		endpoint := rm.getCurrentEndpoint()
		
		rm.logger.Debug("Attempting RPC operation",
			zap.String("endpoint", endpoint),
			zap.String("method", method),
			zap.Int("attempt", i+1))
		
		if err := rm.doRPCCall(ctx, endpoint, method, params, result); err == nil {
			// Success - reset to working endpoint if we had switched
			if rm.currentIdx.Load() != startIdx {
				rm.logger.Info("RPC operation successful, staying on current endpoint")
			}
			return nil
		} else {
			lastErr = err
			rm.logger.Warn("RPC operation failed, trying next endpoint",
				zap.Error(err),
				zap.String("endpoint", endpoint),
				zap.String("method", method))
			
			// Don't switch on the last attempt
			if i < len(rm.endpoints)-1 {
				rm.switchToNext()
			}
		}
	}
	
	return fmt.Errorf("all RPC endpoints failed, last error: %w", lastErr)
}

// doRPCCall performs the actual HTTP JSON-RPC call
func (rm *SolanaRPCManager) doRPCCall(ctx context.Context, endpoint, method string, params interface{}, result interface{}) error {
	request := RPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  method,
		Params:  params,
	}
	
	requestBody, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}
	
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	httpReq.Header.Set("Content-Type", "application/json")
	
	resp, err := rm.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()
	
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	
	var rpcResp RPCResponse
	if err := json.Unmarshal(responseBody, &rpcResp); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}
	
	if rpcResp.Error != nil {
		return fmt.Errorf("RPC error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}
	
	// Marshal and unmarshal result to convert to target type
	resultBytes, err := json.Marshal(rpcResp.Result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}
	
	if err := json.Unmarshal(resultBytes, result); err != nil {
		return fmt.Errorf("failed to unmarshal result: %w", err)
	}
	
	return nil
}

// executeMockOperation provides mock responses for testing
func (rm *SolanaRPCManager) executeMockOperation(method string, result interface{}) error {
	rm.logger.Debug("Executing mock RPC operation for testing", zap.String("method", method))
	
	// Simulate network delay
	time.Sleep(time.Millisecond * 10)
	
	switch method {
	case "getLatestBlockhash":
		if blockResult, ok := result.(*BlockhashResult); ok {
			*blockResult = *rm.getMockBlockhash()
		}
	case "getBalance":
		if balanceResult, ok := result.(*BalanceResult); ok {
			*balanceResult = *rm.getMockBalance("mock_address")
		}
	case "sendTransaction":
		if signature, ok := result.(*string); ok {
			*signature = rm.getMockTransactionSignature()
		}
	case "getSignatureStatuses":
		if statusResult, ok := result.(*SignatureStatusResult); ok {
			*statusResult = *rm.getMockSignatureStatus("mock_signature")
		}
	default:
		rm.logger.Debug("Mock operation not implemented for method", zap.String("method", method))
	}
	
	return nil
}

// GetLatestBlockhash gets the latest blockhash with failover
func (rm *SolanaRPCManager) GetLatestBlockhash(ctx context.Context, commitment string) (*BlockhashResult, error) {
	var result BlockhashResult
	params := map[string]interface{}{
		"commitment": commitment,
	}
	
	err := rm.callRPC(ctx, "getLatestBlockhash", params, &result)
	return &result, err
}

// SendTransaction sends a transaction with failover
func (rm *SolanaRPCManager) SendTransaction(ctx context.Context, transaction string) (string, error) {
	var result string
	params := []interface{}{
		transaction,
		map[string]interface{}{
			"encoding": "base64",
		},
	}
	
	err := rm.callRPC(ctx, "sendTransaction", params, &result)
	return result, err
}

// GetSignatureStatus gets transaction status with failover
func (rm *SolanaRPCManager) GetSignatureStatus(ctx context.Context, signature string) (*SignatureStatusResult, error) {
	var result SignatureStatusResult
	params := []interface{}{
		[]string{signature},
		map[string]interface{}{
			"searchTransactionHistory": true,
		},
	}
	
	err := rm.callRPC(ctx, "getSignatureStatuses", params, &result)
	return &result, err
}

// GetBalance gets account balance with failover
func (rm *SolanaRPCManager) GetBalance(ctx context.Context, address string, commitment string) (*BalanceResult, error) {
	var result BalanceResult
	params := []interface{}{
		address,
		map[string]interface{}{
			"commitment": commitment,
		},
	}
	
	err := rm.callRPC(ctx, "getBalance", params, &result)
	return &result, err
}

// Mock response generators for testing
func (rm *SolanaRPCManager) getMockBlockhash() *BlockhashResult {
	return &BlockhashResult{
		Context: struct {
			Slot uint64 `json:"slot"`
		}{
			Slot: 123456789,
		},
		Value: struct {
			Blockhash            string `json:"blockhash"`
			LastValidBlockHeight uint64 `json:"lastValidBlockHeight"`
		}{
			Blockhash:            "MockBlockhash123456789abcdef",
			LastValidBlockHeight: 123456790,
		},
	}
}

func (rm *SolanaRPCManager) getMockTransactionSignature() string {
	return "MockTransactionSignature123456789abcdef0123456789abcdef"
}

func (rm *SolanaRPCManager) getMockSignatureStatus(_ string) *SignatureStatusResult {
	confirmations := uint64(10)
	return &SignatureStatusResult{
		Context: struct {
			Slot uint64 `json:"slot"`
		}{
			Slot: 123456789,
		},
		Value: []struct {
			Slot               uint64  `json:"slot"`
			Confirmations      *uint64 `json:"confirmations"`
			ConfirmationStatus string  `json:"confirmationStatus"`
			Err                *string `json:"err"`
		}{
			{
				Slot:               123456789,
				Confirmations:      &confirmations,
				ConfirmationStatus: "confirmed",
				Err:                nil,
			},
		},
	}
}

func (rm *SolanaRPCManager) getMockBalance(_ string) *BalanceResult {
	return &BalanceResult{
		Context: struct {
			Slot uint64 `json:"slot"`
		}{
			Slot: 123456789,
		},
		Value: 1000000000, // 1 SOL in lamports
	}
}

// Close closes all RPC connections
func (rm *SolanaRPCManager) Close() error {
	rm.logger.Info("Closing Solana RPC manager")
	// RPC clients don't need explicit closing in gagliardetto/solana-go
	return nil
}