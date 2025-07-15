package env

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// McpSSEClient wraps the official SSE client from mark3labs/mcp-go for testing
type McpSSEClient struct {
	baseURL     string
	sseURL      string
	client      *client.Client
	connected   bool
	initialized bool
}

func NewMcpSSEClient(baseURL string) *McpSSEClient {
	// Extract base URL and construct SSE endpoint
	baseURLWithoutMCP := strings.TrimSuffix(baseURL, "/mcp")
	sseURL := fmt.Sprintf("%s/mcp/sse", baseURLWithoutMCP)
	
	return &McpSSEClient{
		baseURL: baseURL,
		sseURL:  sseURL,
	}
}

func (c *McpSSEClient) Connect(ctx context.Context) error {
	if c.connected {
		return nil
	}

	// Create SSE client using official implementation
	sseClient, err := client.NewSSEMCPClient(c.sseURL)
	if err != nil {
		return fmt.Errorf("failed to create SSE client: %w", err)
	}

	c.client = sseClient
	
	// Start the transport
	if err := c.client.Start(ctx); err != nil {
		return fmt.Errorf("failed to start SSE client transport: %w", err)
	}

	c.connected = true
	return nil
}

func (c *McpSSEClient) Initialize(ctx context.Context) error {
	if c.initialized {
		return nil
	}

	if !c.connected {
		return fmt.Errorf("client not connected")
	}

	// Initialize the MCP session using official client
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	initRequest := mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			Capabilities: mcp.ClientCapabilities{
				Roots: &struct {
					ListChanged bool `json:"listChanged,omitempty"`
				}{
					ListChanged: true,
				},
				Sampling: &struct{}{},
			},
			ClientInfo: mcp.Implementation{
				Name:    "algonius-wallet-test-sse-client",
				Version: "0.1.0",
			},
		},
	}
	
	_, err := c.client.Initialize(ctx, initRequest)

	if err != nil {
		return fmt.Errorf("failed to initialize MCP session: %w", err)
	}

	c.initialized = true
	return nil
}

func (c *McpSSEClient) ListResources(ctx context.Context) (*mcp.ListResourcesResult, error) {
	if !c.initialized {
		return nil, fmt.Errorf("client not initialized")
	}

	result, err := c.client.ListResources(ctx, mcp.ListResourcesRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to list resources: %w", err)
	}

	return result, nil
}

func (c *McpSSEClient) ReadResource(uri string) (*mcp.ReadResourceResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if !c.initialized {
		return nil, fmt.Errorf("client not initialized")
	}

	result, err := c.client.ReadResource(ctx, mcp.ReadResourceRequest{
		Params: mcp.ReadResourceParams{
			URI: uri,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to read resource: %w", err)
	}

	return result, nil
}

func (c *McpSSEClient) Disconnect() error {
	if c.client != nil {
		if err := c.client.Close(); err != nil {
			return fmt.Errorf("failed to close SSE client: %w", err)
		}
	}
	c.connected = false
	c.initialized = false
	return nil
}