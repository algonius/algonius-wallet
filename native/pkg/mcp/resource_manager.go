package mcp

import (
	"context"
	"sync"

	"github.com/mark3labs/mcp-go/server"
)

// ResourceNotifier interface for notifying resource updates
type ResourceNotifier interface {
	NotifyResourceUpdated(uri string)
}

// ResourceManager manages resource subscriptions and notifications
type ResourceManager struct {
	server        *server.MCPServer
	subscriptions map[string]map[string]bool // resourceURI -> sessionID -> subscribed
	mu            sync.RWMutex
}

// NewResourceManager creates a new ResourceManager
func NewResourceManager(server *server.MCPServer) *ResourceManager {
	return &ResourceManager{
		server:        server,
		subscriptions: make(map[string]map[string]bool),
	}
}

// HandleResourceSubscription handles resource subscription
func (rm *ResourceManager) HandleResourceSubscription(ctx context.Context, uri string, sessionID string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if rm.subscriptions[uri] == nil {
		rm.subscriptions[uri] = make(map[string]bool)
	}
	rm.subscriptions[uri][sessionID] = true

	return nil
}

// HandleResourceUnsubscription handles resource unsubscription
func (rm *ResourceManager) HandleResourceUnsubscription(ctx context.Context, uri string, sessionID string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if subs, exists := rm.subscriptions[uri]; exists {
		delete(subs, sessionID)
		if len(subs) == 0 {
			delete(rm.subscriptions, uri)
		}
	}

	return nil
}

// NotifyResourceUpdated notifies all subscribers of a resource update
func (rm *ResourceManager) NotifyResourceUpdated(uri string) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	if subs, exists := rm.subscriptions[uri]; exists {
		for sessionID := range subs {
			_ = rm.server.SendNotificationToSpecificClient(
				sessionID,
				"notifications/resources/updated",
				map[string]interface{}{"uri": uri},
			)
		}
	}
}
