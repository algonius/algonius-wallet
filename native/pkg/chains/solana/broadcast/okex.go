// SPDX-License-Identifier: Apache-2.0
package broadcast

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/algonius/algonius-wallet/native/pkg/clients/okex"
	"github.com/algonius/algonius-wallet/native/pkg/config"
	"go.uber.org/zap"
)

// OKExChannel implements broadcasting via OKEx API
type OKExChannel struct {
	name     string
	enabled  bool
	priority int
	config   *config.OKExConfig
	logger   *zap.Logger
	
	// OKEx API client
	okexClient okex.IOKEXClient
}


// NewOKExChannel creates a new OKEx broadcast channel
func NewOKExChannel(config *config.OKExConfig, logger *zap.Logger) *OKExChannel {
	// Initialize the real OKEx client
	okexClient := okex.Init(config, logger)
	
	return &OKExChannel{
		name:       "okex",
		enabled:    config.Enabled,
		priority:   2, // Lower priority than direct RPC
		config:     config,
		logger:     logger,
		okexClient: okexClient,
	}
}

// GetName returns the channel name
func (o *OKExChannel) GetName() string {
	return o.name
}

// IsEnabled returns whether the channel is enabled
func (o *OKExChannel) IsEnabled() bool {
	return o.enabled && o.config.APIKey != ""
}

// GetPriority returns the channel priority
func (o *OKExChannel) GetPriority() int {
	return o.priority
}

// BroadcastTransaction broadcasts a transaction via OKEx API
func (o *OKExChannel) BroadcastTransaction(ctx context.Context, params *BroadcastParams) (*BroadcastResult, error) {
	startTime := time.Now()
	
	o.logger.Debug("Broadcasting transaction via OKEx API",
		zap.String("signature", params.Signature),
		zap.String("from", params.From),
		zap.String("to", params.To),
		zap.Uint64("amount", params.Amount))
	
	// Handle test mode
	if os.Getenv("RUN_MODE") == "test" {
		return o.broadcastMockTransaction(ctx, params, startTime)
	}
	
	// Create context with timeout
	timeout := time.Duration(o.config.Timeout) * time.Second
	broadcastCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	
	// Prepare OKEx broadcast parameters
	broadcastParams := okex.BroadcastTransactionParams{
		ChainIndex: "501", // Solana chain ID
		Address:    params.From,
		SignedTx:   "0x" + hex.EncodeToString(params.SignedTransaction),
	}
	
	// Add AccountID if available in metadata
	if accountID, ok := params.Metadata["account_id"].(string); ok && accountID != "" {
		broadcastParams.AccountID = accountID
	}
	
	// Broadcast via OKEx client
	response, err := o.okexClient.BroadcastTransaction(broadcastCtx, broadcastParams)
	if err != nil {
		endTime := time.Now()
		return &BroadcastResult{
			Success:   false,
			Channel:   o.name,
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
			Status:    "failed",
			Error:     err.Error(),
		}, err
	}
	
	endTime := time.Now()
	var signature string
	if len(response.Data) > 0 {
		signature = response.Data[0].OrderId // OKEx returns OrderId, not TxID
	}
	
	result := &BroadcastResult{
		Success:      true,
		Signature:    signature,
		Channel:      o.name,
		StartTime:    startTime,
		EndTime:      endTime,
		Duration:     endTime.Sub(startTime),
		Status:       "pending",
		Confirmations: 0,
		Metadata: map[string]any{
			"okex_response_code": response.Code,
			"okex_message":       response.Message,
			"okex_order_id":      signature,
		},
	}
	
	o.logger.Info("Transaction broadcasted successfully via OKEx API",
		zap.String("order_id", signature),
		zap.Duration("duration", result.Duration))
	
	return result, nil
}

// GetTransactionStatus checks transaction status via OKEx API
func (o *OKExChannel) GetTransactionStatus(ctx context.Context, signature string) (*TransactionStatus, error) {
	if os.Getenv("RUN_MODE") == "test" {
		return o.getMockTransactionStatus(signature), nil
	}
	
	// Query OKEx order status
	queryParams := okex.QueryOrdersParams{
		OrderID: signature,
		Limit:   "1",
	}
	
	// Add address if available (OKEx requires either address or accountId)
	if o.config.APIKey != "" {
		// In a real implementation, we might store the address associated with this order
		// For now, we'll try without address and let the API validate
	}
	
	response, err := o.okexClient.GetOrders(ctx, queryParams)
	if err != nil {
		o.logger.Warn("Failed to query OKEx order status",
			zap.String("order_id", signature),
			zap.Error(err))
		return &TransactionStatus{
			Signature:     signature,
			Status:        "unknown",
			Confirmations: 0,
			Error:         err.Error(),
		}, nil
	}
	
	// Parse OKEx response
	status := "unknown"
	confirmations := 0
	var txHash string
	
	if len(response.Data) > 0 {
		order := response.Data[0]
		txHash = order.TxHash
		
		switch order.TxStatus {
		case okex.TxStatusPending:
			status = "pending"
			confirmations = 0
		case okex.TxStatusSuccess:
			status = "confirmed"
			confirmations = 1
		case okex.TxStatusFailed:
			status = "failed"
			confirmations = 0
		}
	}
	
	return &TransactionStatus{
		Signature:     signature,
		Status:        status,
		Confirmations: confirmations,
		Metadata: map[string]any{
			"okex_order_id": signature,
			"tx_hash":       txHash,
			"channel":       "okex",
		},
	}, nil
}

// Close cleans up resources
func (o *OKExChannel) Close() error {
	o.logger.Debug("Closing OKEx broadcast channel")
	return nil
}


// broadcastMockTransaction handles test mode broadcasting
func (o *OKExChannel) broadcastMockTransaction(_ context.Context, params *BroadcastParams, startTime time.Time) (*BroadcastResult, error) {
	// Simulate processing time
	time.Sleep(time.Millisecond * 100)
	
	mockSignature := fmt.Sprintf("MockOKEx_%s_%d", params.From[:8], time.Now().Unix())
	endTime := time.Now()
	
	o.logger.Debug("Mock transaction broadcasted via OKEx API",
		zap.String("signature", mockSignature))
	
	return &BroadcastResult{
		Success:      true,
		Signature:    mockSignature,
		Channel:      o.name,
		StartTime:    startTime,
		EndTime:      endTime,
		Duration:     endTime.Sub(startTime),
		Status:       "pending",
		Confirmations: 0,
		Metadata: map[string]any{
			"mock_mode":      true,
			"test_signature": mockSignature,
			"okex_api":       true,
		},
	}, nil
}

// getMockTransactionStatus returns mock transaction status for testing
func (o *OKExChannel) getMockTransactionStatus(signature string) *TransactionStatus {
	return &TransactionStatus{
		Signature:     signature,
		Status:        "confirmed",
		Confirmations: 1,
		Slot:          123456789,
		BlockTime:     time.Now(),
		Fee:           5000, // 0.000005 SOL
		Metadata: map[string]any{
			"mock_mode": true,
			"channel":   "okex",
		},
	}
}