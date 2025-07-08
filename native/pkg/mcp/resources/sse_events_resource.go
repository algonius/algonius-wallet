// Package resources implements MCP resources for the Algonius Native Host.
package resources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/algonius/algonius-wallet/native/pkg/events"
	"github.com/algonius/algonius-wallet/native/pkg/logger"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// SSEEventsResource implements the events://sse resource for real-time event streaming
type SSEEventsResource struct {
	broadcaster *events.EventBroadcaster
	logger      logger.Logger
}

// NewSSEEventsResource creates a new SSE events resource instance
func NewSSEEventsResource(broadcaster *events.EventBroadcaster, logr logger.Logger) *SSEEventsResource {
	return &SSEEventsResource{
		broadcaster: broadcaster,
		logger:      logr,
	}
}

// GetMeta returns the MCP resource definition for SSE events
func (r *SSEEventsResource) GetMeta() mcp.Resource {
	return mcp.NewResource(
		"events://sse",
		"SSE Events Stream",
		mcp.WithResourceDescription("Real-time event stream for wallet operations, transactions, and status changes"),
		mcp.WithMIMEType("text/event-stream"),
	)
}

// GetHandler returns the resource handler function for SSE events
func (r *SSEEventsResource) GetHandler() server.ResourceHandlerFunc {
	return func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		// For SSE, we return information about available event types
		eventInfo := map[string]interface{}{
			"event_types": []string{
				string(events.EventTypeTransactionConfirmed),
				string(events.EventTypeTransactionPending),
				string(events.EventTypeTransactionFailed),
				string(events.EventTypeBalanceChanged),
				string(events.EventTypeWalletStatusChanged),
				string(events.EventTypeBlockNew),
			},
			"description": "Real-time event stream for wallet operations",
			"usage": map[string]interface{}{
				"connection": "Connect to the SSE endpoint to receive real-time events",
				"format":     "Server-Sent Events (SSE) with JSON payload",
				"endpoint":   "/sse/events",
			},
			"event_schema": map[string]interface{}{
				"id":        "string - unique event identifier",
				"type":      "string - event type (see event_types)",
				"timestamp": "string - ISO 8601 timestamp",
				"chain":     "string - blockchain identifier (optional)",
				"data":      "object - event-specific data",
			},
		}

		data, err := json.MarshalIndent(eventInfo, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal event info: %w", err)
		}

		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      "events://sse",
				MIMEType: "application/json",
				Text:     string(data),
			},
		}, nil
	}
}
