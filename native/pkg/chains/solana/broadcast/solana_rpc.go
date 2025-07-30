// SPDX-License-Identifier: Apache-2.0
package broadcast

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/algonius/algonius-wallet/native/pkg/config"
	"go.uber.org/zap"
)

// SolanaRPCChannel implements broadcasting via direct Solana RPC calls
type SolanaRPCChannel struct {
	name     string
	enabled  bool
	priority int
	config   *config.SolanaChainConfig
	logger   *zap.Logger
	
	// HTTP client for RPC calls
	rpcClient *SolanaRPCClient
}

// SolanaRPCClient is a simplified RPC client for transaction broadcasting
type SolanaRPCClient struct {
	endpoints []string
	timeout   time.Duration
	logger    *zap.Logger
}

// NewSolanaRPCChannel creates a new Solana RPC broadcast channel
func NewSolanaRPCChannel(config *config.SolanaChainConfig, logger *zap.Logger) *SolanaRPCChannel {
	rpcClient := &SolanaRPCClient{
		endpoints: config.RPCEndpoints,
		timeout:   30 * time.Second,
		logger:    logger,
	}
	
	return &SolanaRPCChannel{
		name:      "solana-rpc",
		enabled:   config.Enabled,
		priority:  1, // Highest priority for direct RPC
		config:    config,
		logger:    logger,
		rpcClient: rpcClient,
	}
}

// GetName returns the channel name
func (s *SolanaRPCChannel) GetName() string {
	return s.name
}

// IsEnabled returns whether the channel is enabled
func (s *SolanaRPCChannel) IsEnabled() bool {
	return s.enabled
}

// GetPriority returns the channel priority
func (s *SolanaRPCChannel) GetPriority() int {
	return s.priority
}

// BroadcastTransaction broadcasts a transaction via Solana RPC
func (s *SolanaRPCChannel) BroadcastTransaction(ctx context.Context, params *BroadcastParams) (*BroadcastResult, error) {
	startTime := time.Now()
	
	s.logger.Debug("Broadcasting transaction via Solana RPC",
		zap.String("signature", params.Signature),
		zap.String("from", params.From),
		zap.String("to", params.To),
		zap.Uint64("amount", params.Amount))
	
	// Handle test mode
	if os.Getenv("RUN_MODE") == "test" {
		return s.broadcastMockTransaction(ctx, params, startTime)
	}
	
	// Create context with timeout
	broadcastCtx, cancel := context.WithTimeout(ctx, params.Timeout)
	if params.Timeout == 0 {
		broadcastCtx, cancel = context.WithTimeout(ctx, 30*time.Second)
	}
	defer cancel()
	
	// Broadcast via RPC client
	signature, err := s.rpcClient.SendTransaction(broadcastCtx, params)
	if err != nil {
		endTime := time.Now()
		return &BroadcastResult{
			Success:   false,
			Channel:   s.name,
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
			Status:    "failed",
			Error:     err.Error(),
		}, err
	}
	
	endTime := time.Now()
	result := &BroadcastResult{
		Success:      true,
		Signature:    signature,
		Channel:      s.name,
		StartTime:    startTime,
		EndTime:      endTime,
		Duration:     endTime.Sub(startTime),
		Status:       "pending",
		Confirmations: 0,
		Metadata: map[string]interface{}{
			"rpc_endpoints": s.config.RPCEndpoints,
			"commitment":   s.config.Commitment,
		},
	}
	
	s.logger.Info("Transaction broadcasted successfully via Solana RPC",
		zap.String("signature", signature),
		zap.Duration("duration", result.Duration))
	
	return result, nil
}

// GetTransactionStatus checks transaction status via RPC
func (s *SolanaRPCChannel) GetTransactionStatus(ctx context.Context, signature string) (*TransactionStatus, error) {
	if os.Getenv("RUN_MODE") == "test" {
		return s.getMockTransactionStatus(signature), nil
	}
	
	// TODO: Implement actual RPC status checking
	// This would query getSignatureStatuses RPC method
	return &TransactionStatus{
		Signature:     signature,
		Status:        "pending",
		Confirmations: 0,
	}, nil
}

// Close cleans up resources
func (s *SolanaRPCChannel) Close() error {
	s.logger.Debug("Closing Solana RPC broadcast channel")
	return nil
}

// SendTransaction sends a transaction via RPC
func (client *SolanaRPCClient) SendTransaction(ctx context.Context, params *BroadcastParams) (string, error) {
	if len(client.endpoints) == 0 {
		return "", fmt.Errorf("no RPC endpoints configured")
	}
	
	// Try each endpoint until one succeeds
	var lastErr error
	for _, endpoint := range client.endpoints {
		client.logger.Debug("Attempting to send transaction via RPC",
			zap.String("endpoint", endpoint))
		
		signature, err := client.sendToEndpoint(ctx, endpoint, params)
		if err == nil {
			return signature, nil
		}
		
		lastErr = err
		client.logger.Warn("RPC endpoint failed, trying next",
			zap.String("endpoint", endpoint),
			zap.Error(err))
	}
	
	return "", fmt.Errorf("all RPC endpoints failed: %w", lastErr)
}

// sendToEndpoint sends transaction to a specific RPC endpoint
func (client *SolanaRPCClient) sendToEndpoint(_ context.Context, endpoint string, params *BroadcastParams) (string, error) {
	// In a real implementation, this would:
	// 1. Create HTTP request to RPC endpoint
	// 2. Call sendTransaction method with signed transaction
	// 3. Parse response and return signature
	
	// For now, simulate the RPC call
	time.Sleep(time.Millisecond * 100) // Simulate network delay
	
	// Generate a mock signature based on the transaction
	mockSignature := fmt.Sprintf("SolanaRPC_%s_%d", params.From[:8], time.Now().Unix())
	
	client.logger.Debug("Mock transaction sent via RPC",
		zap.String("endpoint", endpoint),
		zap.String("signature", mockSignature))
	
	return mockSignature, nil
}

// broadcastMockTransaction handles test mode broadcasting
func (s *SolanaRPCChannel) broadcastMockTransaction(_ context.Context, params *BroadcastParams, startTime time.Time) (*BroadcastResult, error) {
	// Simulate processing time
	time.Sleep(time.Millisecond * 50)
	
	mockSignature := fmt.Sprintf("MockSolanaRPC_%s_%d", params.From[:8], time.Now().Unix())
	endTime := time.Now()
	
	s.logger.Debug("Mock transaction broadcasted via Solana RPC",
		zap.String("signature", mockSignature))
	
	return &BroadcastResult{
		Success:      true,
		Signature:    mockSignature,
		Channel:      s.name,
		StartTime:    startTime,
		EndTime:      endTime,
		Duration:     endTime.Sub(startTime),
		Status:       "pending",
		Confirmations: 0,
		Metadata: map[string]any{
			"mock_mode": true,
			"test_signature": mockSignature,
		},
	}, nil
}

// getMockTransactionStatus returns mock transaction status for testing
func (s *SolanaRPCChannel) getMockTransactionStatus(signature string) *TransactionStatus {
	return &TransactionStatus{
		Signature:     signature,
		Status:        "confirmed",
		Confirmations: 1,
		Slot:          123456789,
		BlockTime:     time.Now(),
		Fee:           5000, // 0.000005 SOL
		Metadata: map[string]any{
			"mock_mode": true,
		},
	}
}