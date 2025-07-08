// Package events provides event broadcasting functionality for real-time notifications.
package events

import (
	"sync"
	"time"

	"github.com/algonius/algonius-wallet/native/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// EventType defines the type of events that can be broadcasted
type EventType string

const (
	EventTypeTransactionConfirmed EventType = "transaction_confirmed"
	EventTypeTransactionPending   EventType = "transaction_pending"
	EventTypeTransactionFailed    EventType = "transaction_failed"
	EventTypeBalanceChanged       EventType = "balance_changed"
	EventTypeWalletStatusChanged  EventType = "wallet_status_changed"
	EventTypeBlockNew             EventType = "block_new"
)

// Event represents a single event to be broadcasted
type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Chain     string                 `json:"chain,omitempty"`
	Data      map[string]interface{} `json:"data"`
}

// EventBroadcaster manages event broadcasting to multiple sessions
type EventBroadcaster struct {
	sessions map[string]chan<- *Event
	mu       sync.RWMutex
	logger   logger.Logger
}

// NewEventBroadcaster creates a new event broadcaster instance
func NewEventBroadcaster(logr logger.Logger) *EventBroadcaster {
	return &EventBroadcaster{
		sessions: make(map[string]chan<- *Event),
		logger:   logr,
	}
}

// RegisterSession registers a new session for event broadcasting
func (eb *EventBroadcaster) RegisterSession(sessionID string, eventCh chan<- *Event) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.sessions[sessionID] = eventCh

	if eb.logger != nil {
		eb.logger.Info("Session registered for event broadcasting", zap.String("sessionID", sessionID))
	}
}

// UnregisterSession removes a session from event broadcasting
func (eb *EventBroadcaster) UnregisterSession(sessionID string) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	delete(eb.sessions, sessionID)

	if eb.logger != nil {
		eb.logger.Info("Session unregistered from event broadcasting", zap.String("sessionID", sessionID))
	}
}

// BroadcastEvent sends an event to all registered sessions
func (eb *EventBroadcaster) BroadcastEvent(event *Event) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	if len(eb.sessions) == 0 {
		if eb.logger != nil {
			eb.logger.Debug("No sessions registered for event broadcasting",
				zap.String("eventType", string(event.Type)))
		}
		return
	}

	successCount := 0
	for sessionID, ch := range eb.sessions {
		select {
		case ch <- event:
			successCount++
		default:
			// Channel is full or closed, log warning
			if eb.logger != nil {
				eb.logger.Warn("Failed to send event to session",
					zap.String("sessionID", sessionID),
					zap.String("eventType", string(event.Type)),
					zap.String("eventID", event.ID))
			}
		}
	}

	if eb.logger != nil {
		eb.logger.Info("Event broadcasted",
			zap.String("eventType", string(event.Type)),
			zap.String("eventID", event.ID),
			zap.Int("totalSessions", len(eb.sessions)),
			zap.Int("successfulDeliveries", successCount))
	}
}

// CreateTransactionEvent creates a transaction-related event
func CreateTransactionEvent(eventType EventType, chain, txHash, from, to, amount, token string, extraData map[string]interface{}) *Event {
	data := map[string]interface{}{
		"tx_hash": txHash,
		"from":    from,
		"to":      to,
		"amount":  amount,
	}

	if token != "" {
		data["token"] = token
	}

	// Merge extra data
	for k, v := range extraData {
		data[k] = v
	}

	return &Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Timestamp: time.Now(),
		Chain:     chain,
		Data:      data,
	}
}

// CreateBalanceChangedEvent creates a balance changed event
func CreateBalanceChangedEvent(chain, address, newBalance string, extraData map[string]interface{}) *Event {
	data := map[string]interface{}{
		"address":     address,
		"new_balance": newBalance,
	}

	// Merge extra data
	for k, v := range extraData {
		data[k] = v
	}

	return &Event{
		ID:        uuid.New().String(),
		Type:      EventTypeBalanceChanged,
		Timestamp: time.Now(),
		Chain:     chain,
		Data:      data,
	}
}

// CreateWalletStatusEvent creates a wallet status changed event
func CreateWalletStatusEvent(status string, extraData map[string]interface{}) *Event {
	data := map[string]interface{}{
		"status": status,
	}

	// Merge extra data
	for k, v := range extraData {
		data[k] = v
	}

	return &Event{
		ID:        uuid.New().String(),
		Type:      EventTypeWalletStatusChanged,
		Timestamp: time.Now(),
		Data:      data,
	}
}
