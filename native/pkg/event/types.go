// Package event provides event types and broadcasting functionality for the Algonius Native Host.
package event

import (
	"time"
)

// Event represents a generic event that can be broadcasted to AI Agents
type Event struct {
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// NewEvent creates a new event with the current timestamp
func NewEvent(eventType string, data map[string]interface{}) *Event {
	return &Event{
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
	}
}

// EventType constants for different types of events
const (
	EventTypeTransactionConfirmationNeeded = "transaction_confirmation_needed"
	EventTypeSignatureConfirmationNeeded   = "signature_confirmation_needed"
	EventTypeTransactionConfirmed          = "transaction_confirmed"
	EventTypeTransactionRejected           = "transaction_rejected"
	EventTypeTransactionError              = "transaction_error"
	EventTypeBalanceUpdated                = "balance_updated"
	EventTypeWalletConnected               = "wallet_connected"
	EventTypeWalletDisconnected            = "wallet_disconnected"
)