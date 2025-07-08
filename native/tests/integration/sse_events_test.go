package integration

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/algonius/algonius-wallet/native/pkg/events"
	"github.com/algonius/algonius-wallet/native/pkg/logger"
	"github.com/algonius/algonius-wallet/native/pkg/mcp"
	"github.com/algonius/algonius-wallet/native/pkg/mcp/resources"
	"github.com/algonius/algonius-wallet/native/pkg/mcp/tools"
	"github.com/algonius/algonius-wallet/native/pkg/wallet"
	"github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestSSEEventsBroadcasting(t *testing.T) {
	// Skip test if INTEGRATION_TESTS is not enabled
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test. Set INTEGRATION_TESTS=true to enable.")
	}

	// Create logger
	logr, err := logger.NewLogger("sse_test")
	require.NoError(t, err)

	// Create event broadcaster
	broadcaster := events.NewEventBroadcaster(logr)

	// Create wallet manager
	walletManager := wallet.NewWalletManager()

	// Create MCP server with SSE support
	s := server.NewMCPServer("Test SSE Server", "1.0.0")

	// Register SSE events resource
	mcp.RegisterResource(s, resources.NewSSEEventsResource(broadcaster, logr))

	// Register a simple tool for testing
	createWalletTool := tools.NewCreateWalletTool(walletManager)
	mcp.RegisterTool(s, createWalletTool)

	// Start SSE server
	sseServer := server.NewSSEServer(s)
	port := ":0" // Use random available port

	go func() {
		if err := sseServer.Start(port); err != nil {
			logr.Error("SSE server failed to start", zap.Error(err))
		}
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	t.Run("EventBroadcasterBasicFunctionality", func(t *testing.T) {
		// Test event creation and broadcasting
		eventCh := make(chan *events.Event, 10)
		sessionID := "test-session-1"

		// Register session
		broadcaster.RegisterSession(sessionID, eventCh)

		// Create and broadcast a test event
		testEvent := events.CreateTransactionEvent(
			events.EventTypeTransactionConfirmed,
			"ethereum",
			"0x1234567890abcdef",
			"0xfrom",
			"0xto",
			"1.5",
			"ETH",
			map[string]interface{}{
				"gas_used": "21000",
				"block":    "12345",
			},
		)

		broadcaster.BroadcastEvent(testEvent)

		// Verify event was received
		select {
		case receivedEvent := <-eventCh:
			assert.Equal(t, testEvent.ID, receivedEvent.ID)
			assert.Equal(t, testEvent.Type, receivedEvent.Type)
			assert.Equal(t, testEvent.Chain, receivedEvent.Chain)
			assert.Equal(t, "0x1234567890abcdef", receivedEvent.Data["tx_hash"])
			assert.Equal(t, "21000", receivedEvent.Data["gas_used"])
		case <-time.After(1 * time.Second):
			t.Fatal("Event not received within timeout")
		}

		// Unregister session
		broadcaster.UnregisterSession(sessionID)

		// Test that events are not received after unregistering
		testEvent2 := events.CreateBalanceChangedEvent(
			"ethereum",
			"0xtest",
			"10.5",
			nil,
		)
		broadcaster.BroadcastEvent(testEvent2)

		select {
		case <-eventCh:
			t.Fatal("Event received after session was unregistered")
		case <-time.After(100 * time.Millisecond):
			// Expected - no event should be received
		}
	})

	t.Run("MultipleSessionsBroadcasting", func(t *testing.T) {
		// Test broadcasting to multiple sessions
		eventCh1 := make(chan *events.Event, 10)
		eventCh2 := make(chan *events.Event, 10)
		eventCh3 := make(chan *events.Event, 10)

		sessionIDs := []string{"session-1", "session-2", "session-3"}
		channels := []chan *events.Event{eventCh1, eventCh2, eventCh3}

		// Register all sessions
		for i, sessionID := range sessionIDs {
			broadcaster.RegisterSession(sessionID, channels[i])
		}

		// Create and broadcast a test event
		testEvent := events.CreateWalletStatusEvent(
			"connected",
			map[string]interface{}{
				"wallet_address": "0x1234567890abcdef",
				"network":        "mainnet",
			},
		)

		broadcaster.BroadcastEvent(testEvent)

		// Verify all sessions received the event
		for i, ch := range channels {
			select {
			case receivedEvent := <-ch:
				assert.Equal(t, testEvent.ID, receivedEvent.ID)
				assert.Equal(t, testEvent.Type, receivedEvent.Type)
				assert.Equal(t, "connected", receivedEvent.Data["status"])
				assert.Equal(t, "0x1234567890abcdef", receivedEvent.Data["wallet_address"])
			case <-time.After(1 * time.Second):
				t.Fatalf("Event not received by session %d within timeout", i)
			}
		}

		// Cleanup
		for _, sessionID := range sessionIDs {
			broadcaster.UnregisterSession(sessionID)
		}
	})

	t.Run("EventTypesValidation", func(t *testing.T) {
		// Test all event types
		eventCh := make(chan *events.Event, 20)
		sessionID := "validation-session"
		broadcaster.RegisterSession(sessionID, eventCh)

		// Test transaction events
		txEvent := events.CreateTransactionEvent(
			events.EventTypeTransactionPending,
			"polygon",
			"0xpending123",
			"0xfrom",
			"0xto",
			"0.1",
			"MATIC",
			map[string]interface{}{"nonce": 42},
		)
		broadcaster.BroadcastEvent(txEvent)

		// Test balance changed event
		balanceEvent := events.CreateBalanceChangedEvent(
			"ethereum",
			"0xuser",
			"150.75",
			map[string]interface{}{"previous_balance": "100.25"},
		)
		broadcaster.BroadcastEvent(balanceEvent)

		// Test wallet status event
		statusEvent := events.CreateWalletStatusEvent(
			"disconnected",
			map[string]interface{}{"reason": "user_action"},
		)
		broadcaster.BroadcastEvent(statusEvent)

		// Verify all events
		receivedEvents := make([]*events.Event, 0, 3)
		for i := 0; i < 3; i++ {
			select {
			case event := <-eventCh:
				receivedEvents = append(receivedEvents, event)
			case <-time.After(1 * time.Second):
				t.Fatalf("Event %d not received within timeout", i)
			}
		}

		// Verify event types
		eventTypes := make(map[events.EventType]bool)
		for _, event := range receivedEvents {
			eventTypes[event.Type] = true
		}

		assert.True(t, eventTypes[events.EventTypeTransactionPending])
		assert.True(t, eventTypes[events.EventTypeBalanceChanged])
		assert.True(t, eventTypes[events.EventTypeWalletStatusChanged])

		broadcaster.UnregisterSession(sessionID)
	})
}

func TestSSEResourceHandler(t *testing.T) {
	// Skip test if INTEGRATION_TESTS is not enabled
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test. Set INTEGRATION_TESTS=true to enable.")
	}

	// Create logger
	logr, err := logger.NewLogger("sse_resource_test")
	require.NoError(t, err)

	// Create event broadcaster
	broadcaster := events.NewEventBroadcaster(logr)

	// Create SSE events resource
	sseResource := resources.NewSSEEventsResource(broadcaster, logr)

	t.Run("ResourceMetadata", func(t *testing.T) {
		meta := sseResource.GetMeta()
		assert.Equal(t, "events://sse", meta.URI)
		assert.Equal(t, "SSE Events Stream", meta.Name)
		assert.Contains(t, meta.Description, "Real-time event stream")
		assert.Equal(t, "text/event-stream", meta.MIMEType)
	})

	t.Run("ResourceHandler", func(t *testing.T) {
		handler := sseResource.GetHandler()
		require.NotNil(t, handler)

		// For now, we'll skip the direct handler test since the MCP types are internal
		// and focus on testing the resource metadata which is publicly accessible
		t.Skip("Direct handler testing skipped - requires internal MCP types")
	})
}

// TestSSEEndToEnd tests the complete SSE functionality with HTTP client
func TestSSEEndToEnd(t *testing.T) {
	// Skip test if INTEGRATION_TESTS is not enabled
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test. Set INTEGRATION_TESTS=true to enable.")
	}

	// This test would require starting an actual SSE server
	// and connecting with an HTTP client to test the full flow
	// For now, we'll mark it as a placeholder for manual testing
	t.Skip("End-to-end SSE test requires manual verification with actual server")
}

// Example of how to manually test SSE with curl:
func ExampleSSEManualTest() {
	fmt.Println("Manual SSE testing:")
	fmt.Println("1. Set environment variable: export MCP_SERVER_TYPE=sse")
	fmt.Println("2. Start the native host: ./bin/native-host")
	fmt.Println("3. Connect with curl: curl -N -H 'Accept: text/event-stream' http://localhost:9444/sse/events")
	fmt.Println("4. Trigger wallet operations to see events")
}

// BenchmarkEventBroadcasting benchmarks the event broadcasting performance
func BenchmarkEventBroadcasting(b *testing.B) {
	logr, _ := logger.NewLogger("benchmark")
	broadcaster := events.NewEventBroadcaster(logr)

	// Setup multiple sessions
	numSessions := 100
	channels := make([]chan *events.Event, numSessions)
	for i := 0; i < numSessions; i++ {
		channels[i] = make(chan *events.Event, 1000)
		broadcaster.RegisterSession(fmt.Sprintf("session-%d", i), channels[i])
	}

	// Benchmark event broadcasting
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		event := events.CreateTransactionEvent(
			events.EventTypeTransactionConfirmed,
			"ethereum",
			fmt.Sprintf("0x%d", i),
			"0xfrom",
			"0xto",
			"1.0",
			"ETH",
			nil,
		)
		broadcaster.BroadcastEvent(event)
	}

	// Cleanup
	for i := 0; i < numSessions; i++ {
		broadcaster.UnregisterSession(fmt.Sprintf("session-%d", i))
	}
}
