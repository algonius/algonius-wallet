package integration

import (
	"context"
	"testing"
	"time"

	"github.com/algonius/algonius-wallet/native/tests/integration/env"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

func TestSSETransportEndpoints(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	// Test that the unified server provides both HTTP and SSE endpoints
	// by testing with different client types
	
	t.Run("StreamableHTTPClient", func(t *testing.T) {
		// Test existing HTTP streamable client still works
		client := testEnv.GetMcpClient()
		require.NotNil(t, client, "HTTP MCP client should not be nil")

		require.NoError(t, client.Initialize(ctx), "failed to initialize HTTP MCP client")

		result, err := client.ReadResource("chains://supported")
		require.NoError(t, err, "failed to read resource via HTTP client")
		require.NotNil(t, result, "resource result should not be nil")

		t.Logf("HTTP client result: %+v", result)
	})

	t.Run("PureSSEClient", func(t *testing.T) {
		// Test pure SSE client (like Cline) can connect and work
		baseURL := testEnv.GetBaseURL()
		sseClient := env.NewMcpSSEClient(baseURL)
		require.NotNil(t, sseClient, "SSE client should not be nil")

		// Connect to SSE endpoint
		require.NoError(t, sseClient.Connect(ctx), "failed to connect SSE client")
		defer sseClient.Disconnect()

		// Initialize MCP session
		require.NoError(t, sseClient.Initialize(ctx), "failed to initialize SSE client")

		// Test resource reading via SSE
		result, err := sseClient.ReadResource("chains://supported")
		require.NoError(t, err, "failed to read resource via SSE client")
		require.NotNil(t, result, "SSE resource result should not be nil")

		t.Logf("SSE client result: %+v", result)

		// Verify the content is the same as HTTP client
		require.True(t, len(result.Contents) > 0, "resource should have content")
		
		// Cast to TextResourceContents to access URI and Text
		if textContent, ok := result.Contents[0].(mcp.TextResourceContents); ok {
			require.Equal(t, "chains://supported", textContent.URI, "URI should match")
		} else {
			t.Fatalf("Expected TextResourceContents, got %T", result.Contents[0])
		}
	})

	t.Run("SSEResourcesList", func(t *testing.T) {
		// Test listing resources via SSE
		baseURL := testEnv.GetBaseURL()
		sseClient := env.NewMcpSSEClient(baseURL)
		require.NotNil(t, sseClient, "SSE client should not be nil")

		require.NoError(t, sseClient.Connect(ctx), "failed to connect SSE client")
		defer sseClient.Disconnect()

		require.NoError(t, sseClient.Initialize(ctx), "failed to initialize SSE client")

		// List resources
		result, err := sseClient.ListResources(ctx)
		require.NoError(t, err, "failed to list resources via SSE client")
		require.NotNil(t, result, "list resources result should not be nil")

		t.Logf("SSE client resources list: %+v", result)

		// Verify we have some resources
		require.True(t, len(result.Resources) > 0, "should have at least one resource")

		// Look for our expected resource
		found := false
		for _, resource := range result.Resources {
			if resource.URI == "chains://supported" {
				found = true
				break
			}
		}
		require.True(t, found, "should find chains://supported resource")
	})
}

func TestSSETransportCompatibility(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	// Test that both transport types can work simultaneously
	baseURL := testEnv.GetBaseURL()
	
	// Create both clients
	httpClient := testEnv.GetMcpClient()
	sseClient := env.NewMcpSSEClient(baseURL)

	// Initialize both clients
	require.NoError(t, httpClient.Initialize(ctx), "failed to initialize HTTP client")
	
	require.NoError(t, sseClient.Connect(ctx), "failed to connect SSE client")
	defer sseClient.Disconnect()
	require.NoError(t, sseClient.Initialize(ctx), "failed to initialize SSE client")

	// Test that both can read the same resource simultaneously
	httpResult, err := httpClient.ReadResource("chains://supported")
	require.NoError(t, err, "HTTP client failed to read resource")

	sseResult, err := sseClient.ReadResource("chains://supported")
	require.NoError(t, err, "SSE client failed to read resource")

	// Verify both get the same data
	require.Equal(t, len(httpResult.Contents), len(sseResult.Contents), "both clients should get same number of contents")
	
	if len(httpResult.Contents) > 0 && len(sseResult.Contents) > 0 {
		// Cast both to TextResourceContents for comparison
		httpText, httpOk := httpResult.Contents[0].(mcp.TextResourceContents)
		sseText, sseOk := sseResult.Contents[0].(mcp.TextResourceContents)
		
		require.True(t, httpOk, "HTTP result should be TextResourceContents")
		require.True(t, sseOk, "SSE result should be TextResourceContents")
		
		require.Equal(t, httpText.URI, sseText.URI, "both clients should get same URI")
		require.Equal(t, httpText.Text, sseText.Text, "both clients should get same content")
	}

	t.Log("Both HTTP and SSE clients successfully read the same resource with identical results")
}

func TestSSEEndpointDiscovery(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	baseURL := testEnv.GetBaseURL()
	sseClient := env.NewMcpSSEClient(baseURL)

	// Test the endpoint discovery process
	connectCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	require.NoError(t, sseClient.Connect(connectCtx), "failed to connect and discover endpoint")

	t.Logf("Successfully discovered message endpoint via SSE")
}