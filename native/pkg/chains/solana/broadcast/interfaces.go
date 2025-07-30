// SPDX-License-Identifier: Apache-2.0
package broadcast

import (
	"context"
	"fmt"
	"time"

	"github.com/algonius/algonius-wallet/native/pkg/config"
)

// IBroadcastChannel defines the interface for different transaction broadcast methods
type IBroadcastChannel interface {
	// GetName returns the name of the broadcast channel
	GetName() string
	
	// IsEnabled returns whether this channel is enabled
	IsEnabled() bool
	
	// GetPriority returns the priority of this channel (lower = higher priority)
	GetPriority() int
	
	// BroadcastTransaction broadcasts a signed transaction
	BroadcastTransaction(ctx context.Context, params *BroadcastParams) (*BroadcastResult, error)
	
	// GetTransactionStatus checks the status of a broadcasted transaction
	GetTransactionStatus(ctx context.Context, signature string) (*TransactionStatus, error)
	
	// Close cleans up any resources used by the channel
	Close() error
}

// BroadcastParams contains parameters for broadcasting a transaction
type BroadcastParams struct {
	// Transaction data
	SignedTransaction []byte            `json:"signed_transaction"`
	TransactionBase64 string            `json:"transaction_base64"`
	Signature         string            `json:"signature"`
	
	// Transaction metadata
	From     string            `json:"from"`
	To       string            `json:"to"`
	Amount   uint64            `json:"amount"`
	Token    string            `json:"token,omitempty"`
	
	// Broadcasting options
	SkipPreflight  bool   `json:"skip_preflight"`
	MaxRetries     int    `json:"max_retries"`
	PreflightCommitment string `json:"preflight_commitment"`
	
	// Timing
	Timeout time.Duration `json:"timeout"`
	
	// Additional metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// BroadcastResult contains the result of a broadcast operation
type BroadcastResult struct {
	// Result info
	Success    bool   `json:"success"`
	Signature  string `json:"signature"`
	Channel    string `json:"channel"`
	
	// Timing information
	StartTime    time.Time     `json:"start_time"`
	EndTime      time.Time     `json:"end_time"`
	Duration     time.Duration `json:"duration"`
	
	// Status tracking
	Status       string `json:"status"` // pending, confirmed, failed
	Confirmations int   `json:"confirmations"`
	
	// Error information
	Error        string `json:"error,omitempty"`
	ErrorCode    string `json:"error_code,omitempty"`
	
	// Additional metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// TransactionStatus represents the current status of a transaction
type TransactionStatus struct {
	Signature     string    `json:"signature"`
	Status        string    `json:"status"` // pending, confirmed, failed, rejected
	Confirmations int       `json:"confirmations"`
	Slot          uint64    `json:"slot,omitempty"`
	BlockTime     time.Time `json:"block_time,omitempty"`
	Fee           uint64    `json:"fee,omitempty"`
	Error         string    `json:"error,omitempty"`
	
	// Additional chain-specific data
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// BroadcastManager manages multiple broadcast channels with prioritization
type BroadcastManager struct {
	channels map[string]IBroadcastChannel
	config   *config.BroadcastConfig
	enabled  []IBroadcastChannel
}

// NewBroadcastManager creates a new broadcast manager
func NewBroadcastManager(config *config.BroadcastConfig) *BroadcastManager {
	return &BroadcastManager{
		channels: make(map[string]IBroadcastChannel),
		config:   config,
		enabled:  make([]IBroadcastChannel, 0),
	}
}

// RegisterChannel registers a new broadcast channel
func (bm *BroadcastManager) RegisterChannel(channel IBroadcastChannel) {
	bm.channels[channel.GetName()] = channel
	if channel.IsEnabled() {
		bm.enabled = append(bm.enabled, channel)
	}
}

// GetChannel returns a channel by name
func (bm *BroadcastManager) GetChannel(name string) (IBroadcastChannel, bool) {
	channel, exists := bm.channels[name]
	return channel, exists
}

// GetEnabledChannels returns all enabled channels sorted by priority
func (bm *BroadcastManager) GetEnabledChannels() []IBroadcastChannel {
	return bm.enabled
}

// BroadcastWithFallback attempts to broadcast using the primary channel, falling back to others on failure
func (bm *BroadcastManager) BroadcastWithFallback(ctx context.Context, params *BroadcastParams) (*BroadcastResult, error) {
	// Try primary channel first if specified
	if bm.config.Channel != "" {
		if channel, exists := bm.GetChannel(bm.config.Channel); exists && channel.IsEnabled() {
			result, err := channel.BroadcastTransaction(ctx, params)
			if err == nil {
				return result, nil
			}
			// Log error but continue to fallback channels
		}
	}
	
	// Try all enabled channels in priority order
	for _, channel := range bm.enabled {
		result, err := channel.BroadcastTransaction(ctx, params)
		if err == nil {
			return result, nil
		}
		// Continue to next channel on failure
	}
	
	return nil, ErrAllChannelsFailed
}

// Close closes all registered channels
func (bm *BroadcastManager) Close() error {
	for _, channel := range bm.channels {
		if err := channel.Close(); err != nil {
			// Log error but continue closing other channels
		}
	}
	return nil
}

// Common broadcast errors
var (
	ErrAllChannelsFailed    = fmt.Errorf("all broadcast channels failed")
	ErrChannelNotFound      = fmt.Errorf("broadcast channel not found")
	ErrChannelDisabled      = fmt.Errorf("broadcast channel is disabled")
	ErrInvalidTransaction   = fmt.Errorf("invalid transaction data")
	ErrBroadcastTimeout     = fmt.Errorf("broadcast operation timed out")
	ErrTransactionRejected  = fmt.Errorf("transaction was rejected by the network")
)