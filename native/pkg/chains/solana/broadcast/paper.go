// SPDX-License-Identifier: Apache-2.0
package broadcast

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// PaperChannel implements paper trading (simulation) broadcasting
type PaperChannel struct {
	name     string
	enabled  bool
	priority int
	logger   *zap.Logger
	
	// In-memory storage for simulated transactions
	transactions map[string]*PaperTransaction
	mu           sync.RWMutex
}

// PaperTransaction represents a simulated transaction
type PaperTransaction struct {
	Signature     string                 `json:"signature"`
	From          string                 `json:"from"`
	To            string                 `json:"to"`
	Amount        uint64                 `json:"amount"`
	Token         string                 `json:"token"`
	Status        string                 `json:"status"`
	Confirmations int                    `json:"confirmations"`
	CreatedAt     time.Time              `json:"created_at"`
	ConfirmedAt   time.Time              `json:"confirmed_at,omitempty"`
	Fee           uint64                 `json:"fee"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// NewPaperChannel creates a new paper trading broadcast channel
func NewPaperChannel(logger *zap.Logger) *PaperChannel {
	return &PaperChannel{
		name:         "paper",
		enabled:      true,
		priority:     999, // Lowest priority - only used for testing
		logger:       logger,
		transactions: make(map[string]*PaperTransaction),
	}
}

// GetName returns the channel name
func (p *PaperChannel) GetName() string {
	return p.name
}

// IsEnabled returns whether the channel is enabled
func (p *PaperChannel) IsEnabled() bool {
	return p.enabled
}

// GetPriority returns the channel priority
func (p *PaperChannel) GetPriority() int {
	return p.priority
}

// BroadcastTransaction simulates broadcasting a transaction
func (p *PaperChannel) BroadcastTransaction(ctx context.Context, params *BroadcastParams) (*BroadcastResult, error) {
	startTime := time.Now()
	
	p.logger.Debug("Simulating transaction broadcast (paper trading)",
		zap.String("from", params.From),
		zap.String("to", params.To),
		zap.Uint64("amount", params.Amount),
		zap.String("token", params.Token))
	
	// Generate paper transaction signature
	paperSignature := fmt.Sprintf("Paper_%s_%d", params.From[:8], time.Now().UnixNano())
	
	// Create paper transaction record
	paperTx := &PaperTransaction{
		Signature:     paperSignature,
		From:          params.From,
		To:            params.To,
		Amount:        params.Amount,
		Token:         params.Token,
		Status:        "pending",
		Confirmations: 0,
		CreatedAt:     startTime,
		Fee:           5000, // Simulated fee: 0.000005 SOL
		Metadata: map[string]interface{}{
			"paper_trading": true,
			"simulated":     true,
			"original_params": map[string]interface{}{
				"skip_preflight": params.SkipPreflight,
				"max_retries":    params.MaxRetries,
				"timeout":        params.Timeout.String(),
			},
		},
	}
	
	// Store in memory
	p.mu.Lock()
	p.transactions[paperSignature] = paperTx
	p.mu.Unlock()
	
	// Simulate processing time
	time.Sleep(time.Millisecond * 10)
	
	// Auto-confirm after a short delay
	go p.autoConfirmTransaction(paperSignature)
	
	endTime := time.Now()
	result := &BroadcastResult{
		Success:      true,
		Signature:    paperSignature,
		Channel:      p.name,
		StartTime:    startTime,
		EndTime:      endTime,
		Duration:     endTime.Sub(startTime),
		Status:       "pending",
		Confirmations: 0,
		Metadata: map[string]interface{}{
			"paper_trading":    true,
			"simulated_fee":    paperTx.Fee,
			"auto_confirm":     true,
			"confirm_delay_ms": 2000,
		},
	}
	
	p.logger.Info("Paper transaction created successfully",
		zap.String("signature", paperSignature),
		zap.Duration("duration", result.Duration))
	
	return result, nil
}

// GetTransactionStatus returns the status of a paper transaction
func (p *PaperChannel) GetTransactionStatus(ctx context.Context, signature string) (*TransactionStatus, error) {
	p.mu.RLock()
	paperTx, exists := p.transactions[signature]
	p.mu.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("paper transaction not found: %s", signature)
	}
	
	status := &TransactionStatus{
		Signature:     paperTx.Signature,
		Status:        paperTx.Status,
		Confirmations: paperTx.Confirmations,
		Slot:          123456789, // Mock slot
		Fee:           paperTx.Fee,
		Metadata: map[string]interface{}{
			"paper_trading": true,
			"created_at":    paperTx.CreatedAt,
		},
	}
	
	if !paperTx.ConfirmedAt.IsZero() {
		status.BlockTime = paperTx.ConfirmedAt
		status.Metadata["confirmed_at"] = paperTx.ConfirmedAt
	}
	
	return status, nil
}

// Close cleans up resources
func (p *PaperChannel) Close() error {
	p.logger.Debug("Closing paper trading broadcast channel")
	
	p.mu.Lock()
	defer p.mu.Unlock()
	
	// Log final statistics
	totalTxs := len(p.transactions)
	confirmedTxs := 0
	for _, tx := range p.transactions {
		if tx.Status == "confirmed" {
			confirmedTxs++
		}
	}
	
	p.logger.Info("Paper trading session summary",
		zap.Int("total_transactions", totalTxs),
		zap.Int("confirmed_transactions", confirmedTxs))
	
	// Clear transactions
	p.transactions = make(map[string]*PaperTransaction)
	
	return nil
}

// autoConfirmTransaction simulates automatic transaction confirmation
func (p *PaperChannel) autoConfirmTransaction(signature string) {
	// Wait for simulated confirmation time
	time.Sleep(time.Second * 2)
	
	p.mu.Lock()
	defer p.mu.Unlock()
	
	paperTx, exists := p.transactions[signature]
	if !exists {
		return
	}
	
	// Update to confirmed status
	paperTx.Status = "confirmed"
	paperTx.Confirmations = 1
	paperTx.ConfirmedAt = time.Now()
	
	p.logger.Debug("Paper transaction auto-confirmed",
		zap.String("signature", signature),
		zap.Duration("confirmation_time", paperTx.ConfirmedAt.Sub(paperTx.CreatedAt)))
}

// GetAllTransactions returns all paper transactions (for debugging/testing)
func (p *PaperChannel) GetAllTransactions() map[string]*PaperTransaction {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	// Return a copy to avoid race conditions
	result := make(map[string]*PaperTransaction)
	for k, v := range p.transactions {
		txCopy := *v
		result[k] = &txCopy
	}
	
	return result
}

// ClearTransactions clears all stored paper transactions
func (p *PaperChannel) ClearTransactions() {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	p.transactions = make(map[string]*PaperTransaction)
	p.logger.Debug("Cleared all paper transactions")
}

// SetEnabled allows enabling/disabling the paper channel
func (p *PaperChannel) SetEnabled(enabled bool) {
	p.enabled = enabled
	p.logger.Debug("Paper channel enabled status changed",
		zap.Bool("enabled", enabled))
}