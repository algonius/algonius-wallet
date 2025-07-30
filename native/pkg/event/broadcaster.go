// Package event provides event broadcasting functionality for the Algonius Native Host.
package event

import (
	"encoding/json"
	"sync"
	"time"

	"go.uber.org/zap"
)

// EventBroadcaster manages event distribution to connected AI Agents via SSE
type EventBroadcaster struct {
	clients map[string]chan *Event
	mu      sync.RWMutex
	logger  *zap.Logger
}

// NewEventBroadcaster creates a new EventBroadcaster instance
func NewEventBroadcaster(logger *zap.Logger) *EventBroadcaster {
	return &EventBroadcaster{
		clients: make(map[string]chan *Event),
		mu:      sync.RWMutex{},
		logger:  logger,
	}
}

// Subscribe adds a new client to receive events
func (eb *EventBroadcaster) Subscribe(clientID string) chan *Event {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	// Create a buffered channel for the client
	clientChan := make(chan *Event, 100) // Buffer up to 100 events
	eb.clients[clientID] = clientChan

	eb.logger.Info("Client subscribed to events", zap.String("client_id", clientID))
	
	return clientChan
}

// Unsubscribe removes a client from receiving events
func (eb *EventBroadcaster) Unsubscribe(clientID string) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if clientChan, exists := eb.clients[clientID]; exists {
		close(clientChan)
		delete(eb.clients, clientID)
		eb.logger.Info("Client unsubscribed from events", zap.String("client_id", clientID))
	}
}

// Broadcast sends an event to all subscribed clients
func (eb *EventBroadcaster) Broadcast(event *Event) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	if event == nil {
		eb.logger.Warn("Cannot broadcast nil event")
		return
	}

	// Set timestamp if not already set
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	eventJSON, err := json.Marshal(event)
	if err != nil {
		eb.logger.Error("Failed to marshal event", zap.Error(err))
		return
	}

	eb.logger.Info("Broadcasting event", 
		zap.String("type", event.Type),
		zap.Int("subscribers", len(eb.clients)),
		zap.String("event_data", string(eventJSON)))

	// Send to all subscribers
	for clientID, clientChan := range eb.clients {
		select {
		case clientChan <- event:
			// Event sent successfully
		default:
			// Channel is full, log warning but don't block
			eb.logger.Warn("Client channel full, dropping event", 
				zap.String("client_id", clientID),
				zap.String("event_type", event.Type))
		}
	}
}

// GetSubscriberCount returns the number of currently subscribed clients
func (eb *EventBroadcaster) GetSubscriberCount() int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	return len(eb.clients)
}

// BroadcastTransactionConfirmationNeeded broadcasts a transaction confirmation needed event
func (eb *EventBroadcaster) BroadcastTransactionConfirmationNeeded(txHash, chain, from, to, amount, token, origin string) {
	event := NewEvent(EventTypeTransactionConfirmationNeeded, map[string]interface{}{
		"transaction_hash": txHash,
		"chain":           chain,
		"from":            from,
		"to":              to,
		"amount":          amount,
		"token":           token,
		"origin":          origin,
	})
	eb.Broadcast(event)
}

// BroadcastSignatureConfirmationNeeded broadcasts a signature confirmation needed event
func (eb *EventBroadcaster) BroadcastSignatureConfirmationNeeded(requestHash, address, message, origin string) {
	event := NewEvent(EventTypeSignatureConfirmationNeeded, map[string]interface{}{
		"request_hash": requestHash,
		"address":      address,
		"message":      message,
		"origin":       origin,
	})
	eb.Broadcast(event)
}

// BroadcastTransactionConfirmed broadcasts a transaction confirmed event
func (eb *EventBroadcaster) BroadcastTransactionConfirmed(txHash, chain string, confirmations uint64) {
	event := NewEvent(EventTypeTransactionConfirmed, map[string]interface{}{
		"transaction_hash": txHash,
		"chain":           chain,
		"confirmations":   confirmations,
	})
	eb.Broadcast(event)
}

// BroadcastTransactionRejected broadcasts a transaction rejected event
func (eb *EventBroadcaster) BroadcastTransactionRejected(txHash, reason string) {
	event := NewEvent(EventTypeTransactionRejected, map[string]interface{}{
		"transaction_hash": txHash,
		"reason":          reason,
	})
	eb.Broadcast(event)
}
