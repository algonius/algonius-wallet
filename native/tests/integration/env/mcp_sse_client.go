package env

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// McpSSEClient implements a pure SSE client for testing SSE transport compatibility
type McpSSEClient struct {
	baseURL         string
	sseURL          string
	messageURL      string
	sessionID       string
	httpClient      *http.Client
	eventChan       chan SSEEvent
	connected       bool
	initialized     bool
	cancelFunc      context.CancelFunc
}

type SSEEvent struct {
	Event string `json:"event"`
	Data  string `json:"data"`
}

type EndpointData struct {
	URL string `json:"url"`
}

func NewMcpSSEClient(baseURL string) *McpSSEClient {
	// Extract base URL without /mcp suffix for SSE endpoints
	baseURLWithoutMCP := strings.TrimSuffix(baseURL, "/mcp")
	
	return &McpSSEClient{
		baseURL:    baseURL,
		sseURL:     fmt.Sprintf("%s/mcp/sse", baseURLWithoutMCP),
		httpClient: &http.Client{Timeout: 30 * time.Second},
		eventChan:  make(chan SSEEvent, 100),
		connected:  false,
	}
}

func (c *McpSSEClient) Connect(ctx context.Context) error {
	if c.connected {
		return nil
	}

	// Create a cancellable context for SSE connection
	sseCtx, cancel := context.WithCancel(ctx)
	c.cancelFunc = cancel

	// Establish SSE connection
	req, err := http.NewRequestWithContext(sseCtx, "GET", c.sseURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create SSE request: %w", err)
	}

	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to SSE endpoint: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Start reading SSE events
	go c.readSSEEvents(resp.Body)

	// Wait for endpoint event to get message URL
	select {
	case event := <-c.eventChan:
		if event.Event == "endpoint" {
			messageURL := strings.TrimSpace(event.Data)
			// If it's a relative path, make it absolute
			if strings.HasPrefix(messageURL, "/") {
				baseURLWithoutMCP := strings.TrimSuffix(c.baseURL, "/mcp")
				c.messageURL = baseURLWithoutMCP + messageURL
			} else {
				c.messageURL = messageURL
			}
			c.connected = true
			return nil
		}
		return fmt.Errorf("expected endpoint event, got: %s", event.Event)
	case <-time.After(10 * time.Second):
		resp.Body.Close()
		return fmt.Errorf("timeout waiting for endpoint event")
	case <-sseCtx.Done():
		resp.Body.Close()
		return fmt.Errorf("context cancelled")
	}
}

func (c *McpSSEClient) readSSEEvents(body io.ReadCloser) {
	defer body.Close()
	scanner := bufio.NewScanner(body)

	var event SSEEvent
	for scanner.Scan() {
		line := scanner.Text()
		
		if strings.HasPrefix(line, "event:") {
			event.Event = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		} else if strings.HasPrefix(line, "data:") {
			event.Data = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		} else if line == "" && event.Event != "" {
			// End of event, send it
			select {
			case c.eventChan <- event:
			default:
				// Channel full, skip event
			}
			event = SSEEvent{} // Reset for next event
		}
	}
}

func (c *McpSSEClient) Initialize(ctx context.Context) error {
	if c.initialized {
		return nil
	}

	if !c.connected {
		return fmt.Errorf("client not connected")
	}

	// Send initialize request
	initRequest := mcp.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      mcp.NewRequestId(1),
		Request: mcp.Request{
			Method: string(mcp.MethodInitialize),
		},
		Params: map[string]interface{}{
			"protocolVersion": mcp.LATEST_PROTOCOL_VERSION,
			"capabilities": map[string]interface{}{
				"roots": map[string]interface{}{
					"listChanged": true,
				},
				"sampling": map[string]interface{}{},
			},
			"clientInfo": map[string]interface{}{
				"name":    "algonius-wallet-test-sse-client",
				"version": "0.1.0",
			},
		},
	}

	return c.sendRequest(ctx, initRequest)
}

func (c *McpSSEClient) sendRequest(ctx context.Context, request interface{}) error {
	if c.messageURL == "" {
		return fmt.Errorf("message URL not available")
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.messageURL, strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *McpSSEClient) ListResources(ctx context.Context) (*mcp.ListResourcesResult, error) {
	request := mcp.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      mcp.NewRequestId(2),
		Request: mcp.Request{
			Method: string(mcp.MethodResourcesList),
		},
		Params: map[string]interface{}{},
	}

	if err := c.sendRequest(ctx, request); err != nil {
		return nil, err
	}

	// Wait for response via SSE
	select {
	case event := <-c.eventChan:
		if event.Event == "message" {
			// First try to parse as error response
			var errorResponse mcp.JSONRPCError
			if err := json.Unmarshal([]byte(event.Data), &errorResponse); err == nil && errorResponse.Error.Code != 0 {
				return nil, fmt.Errorf("MCP error: %s", errorResponse.Error.Message)
			}

			// Parse as successful response
			var response mcp.JSONRPCResponse
			if err := json.Unmarshal([]byte(event.Data), &response); err != nil {
				return nil, fmt.Errorf("failed to unmarshal response: %w", err)
			}

			// Convert result to JSON bytes then unmarshal
			resultBytes, err := json.Marshal(response.Result)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal result: %w", err)
			}

			var result mcp.ListResourcesResult
			if err := json.Unmarshal(resultBytes, &result); err != nil {
				return nil, fmt.Errorf("failed to unmarshal result: %w", err)
			}

			return &result, nil
		}
		return nil, fmt.Errorf("unexpected event type: %s", event.Event)
	case <-time.After(10 * time.Second):
		return nil, fmt.Errorf("timeout waiting for response")
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (c *McpSSEClient) ReadResource(uri string) (*mcp.ReadResourceResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	request := mcp.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      mcp.NewRequestId(3),
		Request: mcp.Request{
			Method: string(mcp.MethodResourcesRead),
		},
		Params: map[string]interface{}{
			"uri": uri,
		},
	}

	if err := c.sendRequest(ctx, request); err != nil {
		return nil, err
	}

	// Wait for response via SSE
	select {
	case event := <-c.eventChan:
		if event.Event == "message" {
			// First try to parse as error response
			var errorResponse mcp.JSONRPCError
			if err := json.Unmarshal([]byte(event.Data), &errorResponse); err == nil && errorResponse.Error.Code != 0 {
				return nil, fmt.Errorf("MCP error: %s", errorResponse.Error.Message)
			}

			// Parse as successful response
			var response mcp.JSONRPCResponse
			if err := json.Unmarshal([]byte(event.Data), &response); err != nil {
				return nil, fmt.Errorf("failed to unmarshal response: %w", err)
			}

			// Convert result to JSON bytes then unmarshal
			resultBytes, err := json.Marshal(response.Result)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal result: %w", err)
			}

			var result mcp.ReadResourceResult
			if err := json.Unmarshal(resultBytes, &result); err != nil {
				return nil, fmt.Errorf("failed to unmarshal result: %w", err)
			}

			return &result, nil
		}
		return nil, fmt.Errorf("unexpected event type: %s", event.Event)
	case <-time.After(10 * time.Second):
		return nil, fmt.Errorf("timeout waiting for response")
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (c *McpSSEClient) Disconnect() error {
	if c.cancelFunc != nil {
		c.cancelFunc()
	}
	c.connected = false
	c.initialized = false
	return nil
}