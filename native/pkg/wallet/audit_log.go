// SPDX-License-Identifier: Apache-2.0
package wallet

import (
	"time"
	"crypto/rand"
	"encoding/hex"
)

// AuditLogEntry represents an entry in the audit log
type AuditLogEntry struct {
	ID           string    `json:"id"`
	Action       string    `json:"action"`         // "transaction_rejection", "transaction_confirmation", etc.
	Subject      string    `json:"subject"`        // Transaction hash or other identifier
	Details      string    `json:"details"`        // Additional context
	Reason       string    `json:"reason"`         // Action reason
	Timestamp    time.Time `json:"timestamp"`
	Source       string    `json:"source"`         // "ai_agent", "user", "system"
	WalletAddress string   `json:"wallet_address,omitempty"`
}

// AuditLogger handles audit logging operations
type AuditLogger struct {
	// In a real implementation, this would connect to a database or external logging system
	entries []AuditLogEntry
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger() *AuditLogger {
	return &AuditLogger{
		entries: make([]AuditLogEntry, 0),
	}
}

// LogTransactionRejection logs a transaction rejection event
func (al *AuditLogger) LogTransactionRejection(transactionHash, reason, details, walletAddress string) (string, error) {
	id, err := generateAuditLogID()
	if err != nil {
		return "", err
	}

	entry := AuditLogEntry{
		ID:            id,
		Action:        "transaction_rejection",
		Subject:       transactionHash,
		Details:       details,
		Reason:        reason,
		Timestamp:     time.Now().UTC(),
		Source:        "ai_agent",
		WalletAddress: walletAddress,
	}

	// In a real implementation, this would be persisted to a database
	al.entries = append(al.entries, entry)
	
	return id, nil
}

// GetAuditLog retrieves audit log entries
func (al *AuditLogger) GetAuditLog(limit int, offset int) ([]AuditLogEntry, error) {
	if offset >= len(al.entries) {
		return []AuditLogEntry{}, nil
	}

	end := offset + limit
	if end > len(al.entries) {
		end = len(al.entries)
	}

	return al.entries[offset:end], nil
}

// generateAuditLogID generates a unique audit log ID
func generateAuditLogID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "audit_" + hex.EncodeToString(bytes), nil
}