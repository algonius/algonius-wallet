// SPDX-License-Identifier: Apache-2.0
package wallet

import (
	"time"
)

// PendingTransaction represents a transaction that is waiting for confirmation
type PendingTransaction struct {
	Hash                      string    `json:"hash"`
	Chain                     string    `json:"chain"`
	From                      string    `json:"from"`
	To                        string    `json:"to"`
	Amount                    string    `json:"amount"`
	Token                     string    `json:"token"`
	Type                      string    `json:"type"`                        // "transfer", "swap", "contract"
	Status                    string    `json:"status"`                      // "pending", "confirmed", "failed"
	Confirmations             uint64    `json:"confirmations"`
	RequiredConfirmations     uint64    `json:"required_confirmations"`
	BlockNumber               uint64    `json:"block_number,omitempty"`
	Nonce                     uint64    `json:"nonce,omitempty"`
	GasFee                    string    `json:"gas_fee"`
	Priority                  string    `json:"priority"`                    // "low", "medium", "high"
	EstimatedConfirmationTime string    `json:"estimated_confirmation_time"` // Human-readable estimate
	SubmittedAt               time.Time `json:"submitted_at"`
	LastChecked               time.Time `json:"last_checked"`
}

// PendingTransactionFilter defines filtering options for pending transactions
type PendingTransactionFilter struct {
	Chain   string
	Address string
	Type    string
	Limit   int
	Offset  int
}