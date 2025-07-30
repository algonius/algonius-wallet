// SPDX-License-Identifier: Apache-2.0
package broadcast

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/algonius/algonius-wallet/native/pkg/clients/jito"
	"github.com/algonius/algonius-wallet/native/pkg/config"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
	"go.uber.org/zap"
)

// JitoBundleChannel implements broadcasting via Jito bundles with MEV protection
type JitoBundleChannel struct {
	name       string
	enabled    bool
	priority   int
	config     *config.JitoConfig
	logger     *zap.Logger
	rpcClient  *rpc.Client

	// Jito API client
	jitoClient jito.IJitoAPI
	tipAccount solana.PublicKey
}

// NewJitoBundleChannel creates a new Jito bundle broadcast channel
func NewJitoBundleChannel(cfg *config.JitoConfig, rpcEndpoint string, logger *zap.Logger) *JitoBundleChannel {
	// Initialize the real Jito client
	jitoClient := jito.Init(cfg)
	
	// Create RPC client for transaction monitoring
	rpcClient := rpc.New(rpcEndpoint)
	
	return &JitoBundleChannel{
		name:       "jito-bundle",
		enabled:    cfg.Enabled,
		priority:   4, // Lower priority for bundle mode
		config:     cfg,
		logger:     logger,
		rpcClient:  rpcClient,
		jitoClient: jitoClient,
	}
}

// Init initializes the Jito bundle channel by getting a tip account
func (j *JitoBundleChannel) Init(ctx context.Context) error {
	if !j.enabled {
		return nil
	}
	
	// Get random tip account for bundles
	tipAccount, err := j.jitoClient.GetRandomTipAccount()
	if err != nil {
		return fmt.Errorf("failed to get jito tip account: %w", err)
	}

	tipPubKey, err := solana.PublicKeyFromBase58(tipAccount.Address)
	if err != nil {
		return fmt.Errorf("invalid tip account address: %w", err)
	}

	j.tipAccount = tipPubKey
	j.logger.Info("Jito bundle channel initialized",
		zap.String("tip_account", tipAccount.Address))
	
	return nil
}

// GetName returns the channel name
func (j *JitoBundleChannel) GetName() string {
	return j.name
}

// IsEnabled returns whether the channel is enabled
func (j *JitoBundleChannel) IsEnabled() bool {
	return j.enabled
}

// GetPriority returns the channel priority
func (j *JitoBundleChannel) GetPriority() int {
	return j.priority
}

// BroadcastTransaction broadcasts a transaction via Jito bundle
func (j *JitoBundleChannel) BroadcastTransaction(ctx context.Context, params *BroadcastParams) (*BroadcastResult, error) {
	startTime := time.Now()
	
	j.logger.Debug("Broadcasting transaction via Jito bundle",
		zap.String("signature", params.Signature),
		zap.String("from", params.From),
		zap.String("to", params.To),
		zap.Uint64("amount", params.Amount))
	
	// Handle test mode
	if os.Getenv("RUN_MODE") == "test" {
		return j.broadcastMockTransaction(ctx, params, startTime)
	}
	
	// Create main transaction from signed bytes
	mainTx, err := solana.TransactionFromBytes(params.SignedTransaction)
	if err != nil {
		endTime := time.Now()
		return &BroadcastResult{
			Success:   false,
			Channel:   j.name,
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
			Status:    "failed",
			Error:     fmt.Sprintf("failed to decode transaction: %v", err),
		}, err
	}
	
	// Get tip amount (default to base tip if not specified)
	tipAmount := j.config.BaseTipLamports
	if tipAmountMeta, ok := params.Metadata["jito_tip_amount"].(uint64); ok && tipAmountMeta > 0 {
		tipAmount = tipAmountMeta
		if tipAmount > j.config.MaxTipLamports {
			tipAmount = j.config.MaxTipLamports
		}
	}
	
	// Create tip transaction
	tipTx, err := j.createTipTransaction(params, tipAmount, mainTx.Message.RecentBlockhash)
	if err != nil {
		endTime := time.Now()
		return &BroadcastResult{
			Success:   false,
			Channel:   j.name,
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
			Status:    "failed",
			Error:     fmt.Sprintf("failed to create tip transaction: %v", err),
		}, err
	}
	
	// Prepare bundle request (Jito expects [][]string format)
	bundleRequest := [][]string{
		{
			j.encodeTransaction(mainTx),
			j.encodeTransaction(tipTx),
		},
	}
	
	// Send bundle
	j.logger.Debug("Sending Jito bundle", zap.Uint64("tip_amount", tipAmount))
	bundleIdRaw, err := j.jitoClient.SendBundle(bundleRequest)
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
	
	var bundleId string
	if err := json.Unmarshal(bundleIdRaw, &bundleId); err != nil {
		endTime := time.Now()
		return &BroadcastResult{
			Success:   false,
			Channel:   j.name,
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
			Status:    "failed",
			Error:     fmt.Sprintf("failed to unmarshal bundle ID: %v", err),
		}, err
	}
	
	endTime := time.Now()
	
	result := &BroadcastResult{
		Success:      true,
		Signature:    bundleId, // Use bundle ID as signature for tracking
		Channel:      j.name,
		StartTime:    startTime,
		EndTime:      endTime,
		Duration:     endTime.Sub(startTime),
		Status:       "pending",
		Confirmations: 0,
		Metadata: map[string]any{
			"jito_bundle_id": bundleId,
			"tip_amount":     tipAmount,
			"mev_protected":  true,
			"bundle_mode":    true,
		},
	}
	
	j.logger.Info("Bundle submitted successfully via Jito",
		zap.String("bundle_id", bundleId),
		zap.Uint64("tip_amount", tipAmount),
		zap.Duration("duration", result.Duration))
	
	return result, nil
}

// GetTransactionStatus checks bundle status via Jito API
func (j *JitoBundleChannel) GetTransactionStatus(ctx context.Context, bundleId string) (*TransactionStatus, error) {
	if os.Getenv("RUN_MODE") == "test" {
		return j.getMockTransactionStatus(bundleId), nil
	}
	
	// Query bundle status
	statusResponse, err := j.jitoClient.GetBundleStatuses([]string{bundleId})
	if err != nil {
		j.logger.Warn("Failed to get bundle status",
			zap.String("bundle_id", bundleId),
			zap.Error(err))
		return &TransactionStatus{
			Signature:     bundleId,
			Status:        "unknown",
			Confirmations: 0,
			Error:         err.Error(),
		}, nil
	}
	
	if len(statusResponse.Value) == 0 {
		return &TransactionStatus{
			Signature:     bundleId,
			Status:        "pending",
			Confirmations: 0,
		}, nil
	}
	
	bundleStatus := statusResponse.Value[0]
	
	// Map Jito bundle status to our status
	status := "pending"
	confirmations := 0
	var errorStr string
	var txHashes []string
	
	switch bundleStatus.ConfirmationStatus {
	case "processed":
		status = "processed"
		confirmations = 1
	case "confirmed":
		status = "confirmed"
		confirmations = 20
	case "finalized":
		status = "confirmed"
		confirmations = 32
		txHashes = bundleStatus.Transactions
		
		if bundleStatus.Err.Ok != nil {
			status = "failed"
			errorStr = fmt.Sprintf("%v", bundleStatus.Err.Ok)
		}
	default:
		if bundleStatus.Err.Ok != nil {
			status = "failed"
			errorStr = fmt.Sprintf("%v", bundleStatus.Err.Ok)
		}
	}
	
	return &TransactionStatus{
		Signature:     bundleId,
		Status:        status,
		Confirmations: confirmations,
		BlockTime:     time.Now(), // Jito doesn't provide block time
		Error:         errorStr,
		Metadata: map[string]any{
			"jito_bundle_id":        bundleId,
			"confirmation_status":   bundleStatus.ConfirmationStatus,
			"transaction_hashes":    txHashes,
			"mev_protected":         true,
			"bundle_mode":           true,
			"channel":               "jito-bundle",
		},
	}, nil
}

// Close cleans up resources
func (j *JitoBundleChannel) Close() error {
	j.logger.Debug("Closing Jito bundle broadcast channel")
	return nil
}

// createTipTransaction creates a tip transaction for the bundle
func (j *JitoBundleChannel) createTipTransaction(params *BroadcastParams, tipAmount uint64, recentBlockhash solana.Hash) (*solana.Transaction, error) {
	// Parse the owner private key from metadata
	var ownerPrivateKey solana.PrivateKey
	if privKeyBytes, ok := params.Metadata["owner_private_key"].([]byte); ok && len(privKeyBytes) == 64 {
		copy(ownerPrivateKey[:], privKeyBytes)
	} else {
		return nil, fmt.Errorf("owner private key not found in metadata")
	}
	
	// Create tip transaction
	tipTx, err := solana.NewTransaction(
		[]solana.Instruction{
			system.NewTransferInstruction(
				tipAmount,
				ownerPrivateKey.PublicKey(),
				j.tipAccount,
			).Build(),
		},
		recentBlockhash,
		solana.TransactionPayer(ownerPrivateKey.PublicKey()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create tip transaction: %w", err)
	}
	
	// Sign tip transaction
	_, err = tipTx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if ownerPrivateKey.PublicKey().Equals(key) {
			return &ownerPrivateKey
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to sign tip transaction: %w", err)
	}
	
	return tipTx, nil
}

// encodeTransaction encodes a transaction for bundle submission
func (j *JitoBundleChannel) encodeTransaction(tx *solana.Transaction) string {
	serializedTx, err := tx.MarshalBinary()
	if err != nil {
		j.logger.Error("Failed to serialize transaction", zap.Error(err))
		return ""
	}
	
	return base64.StdEncoding.EncodeToString(serializedTx)
}

// broadcastMockTransaction handles test mode broadcasting
func (j *JitoBundleChannel) broadcastMockTransaction(_ context.Context, params *BroadcastParams, startTime time.Time) (*BroadcastResult, error) {
	// Simulate processing time
	time.Sleep(time.Millisecond * 200)
	
	mockBundleId := fmt.Sprintf("MockJitoBundle_%s_%d", params.From[:8], time.Now().Unix())
	endTime := time.Now()
	
	return &BroadcastResult{
		Success:      true,
		Signature:    mockBundleId,
		Channel:      j.name,
		StartTime:    startTime,
		EndTime:      endTime,
		Duration:     endTime.Sub(startTime),
		Status:       "pending",
		Confirmations: 0,
		Metadata: map[string]any{
			"mock_mode":        true,
			"test_bundle_id":   mockBundleId,
			"jito_bundle":      true,
			"mev_protected":    true,
			"bundle_mode":      true,
		},
	}, nil
}

// getMockTransactionStatus returns mock bundle status for testing
func (j *JitoBundleChannel) getMockTransactionStatus(bundleId string) *TransactionStatus {
	return &TransactionStatus{
		Signature:     bundleId,
		Status:        "confirmed",
		Confirmations: 32, // Finalized confirmations
		Slot:          123456789,
		BlockTime:     time.Now(),
		Fee:           5000, // 0.000005 SOL
		Metadata: map[string]any{
			"mock_mode":      true,
			"channel":        "jito-bundle",
			"mev_protected":  true,
			"bundle_mode":    true,
		},
	}
}