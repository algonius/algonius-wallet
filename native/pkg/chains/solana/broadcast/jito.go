// SPDX-License-Identifier: Apache-2.0
package broadcast

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/algonius/algonius-wallet/native/pkg/clients/jito"
	"github.com/algonius/algonius-wallet/native/pkg/config"
	solana "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/mr-tron/base58"
	"go.uber.org/zap"
)

// JitoChannel implements broadcasting via Jito MEV protection
type JitoChannel struct {
	name       string
	enabled    bool
	priority   int
	config     *config.JitoConfig
	logger     *zap.Logger
	rpcClient  *rpc.Client

	// Jito API client
	jitoClient jito.IJitoAPI
}

// NewJitoChannel creates a new Jito broadcast channel
func NewJitoChannel(cfg *config.JitoConfig, rpcEndpoint string, logger *zap.Logger) *JitoChannel {
	// Initialize the real Jito client
	jitoClient := jito.Init(cfg)
	
	// Create RPC client for transaction status monitoring
	rpcClient := rpc.New(rpcEndpoint)
	
	return &JitoChannel{
		name:       "jito",
		enabled:    cfg.Enabled,
		priority:   3, // Lower priority than direct RPC and OKEx
		config:     cfg,
		logger:     logger,
		rpcClient:  rpcClient,
		jitoClient: jitoClient,
	}
}

// GetName returns the channel name
func (j *JitoChannel) GetName() string {
	return j.name
}

// IsEnabled returns whether the channel is enabled
func (j *JitoChannel) IsEnabled() bool {
	return j.enabled
}

// GetPriority returns the channel priority
func (j *JitoChannel) GetPriority() int {
	return j.priority
}

// BroadcastTransaction broadcasts a transaction via Jito
func (j *JitoChannel) BroadcastTransaction(ctx context.Context, params *BroadcastParams) (*BroadcastResult, error) {
	startTime := time.Now()
	
	j.logger.Debug("Broadcasting transaction via Jito",
		zap.String("signature", params.Signature),
		zap.String("from", params.From),
		zap.String("to", params.To),
		zap.Uint64("amount", params.Amount))
	
	// Handle test mode
	if os.Getenv("RUN_MODE") == "test" {
		return j.broadcastMockTransaction(ctx, params, startTime)
	}
	
	// Serialize and encode transaction for Jito
	encodedTx := base58.Encode(params.SignedTransaction)
	
	// Send transaction via Jito
	txRequest := []string{encodedTx}
	result, err := j.jitoClient.SendTxn(txRequest, false)
	if err != nil {
		endTime := time.Now()
		return &BroadcastResult{
			Success:   false,
			Channel:   j.name,
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
			Status:    "failed",
			Error:     err.Error(),
		}, err
	}
	
	// Parse signature from response
	signature := strings.Trim(string(result), "\"")
	endTime := time.Now()
	
	broadcastResult := &BroadcastResult{
		Success:      true,
		Signature:    signature,
		Channel:      j.name,
		StartTime:    startTime,
		EndTime:      endTime,
		Duration:     endTime.Sub(startTime),
		Status:       "pending",
		Confirmations: 0,
		Metadata: map[string]any{
			"jito_broadcast": true,
			"mev_protected":  true,
		},
	}
	
	j.logger.Info("Transaction broadcasted successfully via Jito",
		zap.String("signature", signature),
		zap.Duration("duration", broadcastResult.Duration))
	
	return broadcastResult, nil
}

// GetTransactionStatus checks transaction status via Solana RPC
func (j *JitoChannel) GetTransactionStatus(ctx context.Context, signature string) (*TransactionStatus, error) {
	if os.Getenv("RUN_MODE") == "test" {
		return j.getMockTransactionStatus(signature), nil
	}
	
	// Parse signature
	sig, err := base58.Decode(signature)
	if err != nil {
		return &TransactionStatus{
			Signature:     signature,
			Status:        "unknown",
			Confirmations: 0,
			Error:         fmt.Sprintf("invalid signature format: %v", err),
		}, nil
	}
	
	var solSig solana.Signature
	copy(solSig[:], sig)
	
	// Query transaction status via RPC
	status, err := j.rpcClient.GetSignatureStatuses(ctx, true, solSig)
	if err != nil {
		j.logger.Warn("Failed to get signature status",
			zap.String("signature", signature),
			zap.Error(err))
		return &TransactionStatus{
			Signature:     signature,
			Status:        "unknown",
			Confirmations: 0,
			Error:         err.Error(),
		}, nil
	}
	
	if len(status.Value) == 0 || status.Value[0] == nil {
		return &TransactionStatus{
			Signature:     signature,
			Status:        "pending",
			Confirmations: 0,
		}, nil
	}
	
	statusValue := status.Value[0]
	
	// Determine status
	txStatus := "pending"
	confirmations := 0
	var errorStr string
	
	if statusValue.Err != nil {
		txStatus = "failed"
		errorStr = fmt.Sprintf("%v", statusValue.Err)
	} else if statusValue.Confirmations != nil {
		confirmations = int(*statusValue.Confirmations)
		if confirmations >= 20 {
			txStatus = "confirmed"
		}
	}
	
	return &TransactionStatus{
		Signature:     signature,
		Status:        txStatus,
		Confirmations: confirmations,
		Slot:          statusValue.Slot,
		BlockTime:     time.Now(), // BlockTime not available in signature status
		Error:         errorStr,
		Metadata: map[string]any{
			"jito_broadcast": true,
			"mev_protected":  true,
			"channel":        "jito",
		},
	}, nil
}

// Close cleans up resources
func (j *JitoChannel) Close() error {
	j.logger.Debug("Closing Jito broadcast channel")
	return nil
}

// broadcastMockTransaction handles test mode broadcasting
func (j *JitoChannel) broadcastMockTransaction(_ context.Context, params *BroadcastParams, startTime time.Time) (*BroadcastResult, error) {
	// Simulate processing time
	time.Sleep(time.Millisecond * 150)
	
	mockSignature := fmt.Sprintf("MockJito_%s_%d", params.From[:8], time.Now().Unix())
	endTime := time.Now()
	
	return &BroadcastResult{
		Success:      true,
		Signature:    mockSignature,
		Channel:      j.name,
		StartTime:    startTime,
		EndTime:      endTime,
		Duration:     endTime.Sub(startTime),
		Status:       "pending",
		Confirmations: 0,
		Metadata: map[string]any{
			"mock_mode":      true,
			"test_signature": mockSignature,
			"jito_api":       true,
			"mev_protected":  true,
		},
	}, nil
}

// getMockTransactionStatus returns mock transaction status for testing
func (j *JitoChannel) getMockTransactionStatus(signature string) *TransactionStatus {
	return &TransactionStatus{
		Signature:     signature,
		Status:        "confirmed",
		Confirmations: 25, // High confirmations for Jito
		Slot:          123456789,
		BlockTime:     time.Now(),
		Fee:           5000, // 0.000005 SOL
		Metadata: map[string]any{
			"mock_mode":      true,
			"channel":        "jito",
			"mev_protected":  true,
		},
	}
}