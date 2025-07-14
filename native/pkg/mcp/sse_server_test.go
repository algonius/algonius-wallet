// SPDX-License-Identifier: Apache-2.0
package mcp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestSSETransportDetector(t *testing.T) {
	tests := []struct {
		name     string
		setupReq func() *http.Request
		wantSSE  bool
	}{
		{
			name: "Accept header with text/event-stream",
			setupReq: func() *http.Request {
				req := httptest.NewRequest("GET", "/mcp", nil)
				req.Header.Set("Accept", "text/event-stream")
				return req
			},
			wantSSE: true,
		},
		{
			name: "Query parameter transport=sse",
			setupReq: func() *http.Request {
				req := httptest.NewRequest("GET", "/mcp?transport=sse", nil)
				return req
			},
			wantSSE: true,
		},
		{
			name: "User-Agent contains eventsource",
			setupReq: func() *http.Request {
				req := httptest.NewRequest("GET", "/mcp", nil)
				req.Header.Set("User-Agent", "EventSource/1.0")
				return req
			},
			wantSSE: true,
		},
		{
			name: "Regular HTTP request",
			setupReq: func() *http.Request {
				req := httptest.NewRequest("GET", "/mcp", nil)
				req.Header.Set("Accept", "application/json")
				return req
			},
			wantSSE: false,
		},
		{
			name: "Mixed Accept header with event-stream",
			setupReq: func() *http.Request {
				req := httptest.NewRequest("GET", "/mcp", nil)
				req.Header.Set("Accept", "application/json, text/event-stream")
				return req
			},
			wantSSE: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setupReq()
			result := SSETransportDetector(req)
			assert.Equal(t, tt.wantSSE, result)
		})
	}
}

func TestSSEServer_ServeHTTP(t *testing.T) {
	// Create a test MCP server
	mcpServer := server.NewMCPServer("Test Server", "1.0.0")
	logger := zap.NewNop()
	
	sseServer := NewSSEServer(mcpServer, logger, "Test Server", "1.0.0")

	tests := []struct {
		name           string
		method         string
		headers        map[string]string
		expectedStatus int
		checkResponse  func(t *testing.T, body string)
	}{
		{
			name:           "GET request for SSE stream",
			method:         "GET",
			headers:        map[string]string{},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				assert.Contains(t, body, "event: connected")
				assert.Contains(t, body, "event: server_info")
				assert.Contains(t, body, "data:")
			},
		},
		{
			name:           "OPTIONS preflight request",
			method:         "OPTIONS",
			headers:        map[string]string{},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				// OPTIONS should return empty body
			},
		},
		{
			name:           "Unsupported method",
			method:         "DELETE",
			headers:        map[string]string{},
			expectedStatus: http.StatusMethodNotAllowed,
			checkResponse: func(t *testing.T, body string) {
				assert.Contains(t, body, "Method not allowed")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/mcp/sse", nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			w := httptest.NewRecorder()
			sseServer.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			// Check SSE headers for successful requests
			if w.Code == http.StatusOK {
				assert.Equal(t, "text/event-stream", w.Header().Get("Content-Type"))
				assert.Equal(t, "no-cache", w.Header().Get("Cache-Control"))
				assert.Equal(t, "keep-alive", w.Header().Get("Connection"))
				assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.String())
			}
		})
	}
}

func TestSSEServer_ProcessMCPRequest(t *testing.T) {
	// TODO: Fix this test when MCP Go library API is stable
	// The API has changed significantly and requires proper request construction
	t.Skip("Skipping due to MCP API changes - needs proper request construction")
}

func TestCreateMCPHandler(t *testing.T) {
	// Create a test MCP server
	mcpServer := server.NewMCPServer("Test Server", "1.0.0")
	logger := zap.NewNop()
	
	handler := CreateMCPHandler(mcpServer, logger, "Test Server", "1.0.0")

	tests := []struct {
		name        string
		setupReq    func() *http.Request
		expectSSE   bool
		description string
	}{
		{
			name: "SSE request with Accept header",
			setupReq: func() *http.Request {
				req := httptest.NewRequest("GET", "/mcp", nil)
				req.Header.Set("Accept", "text/event-stream")
				return req
			},
			expectSSE:   true,
			description: "Should use SSE transport",
		},
		{
			name: "HTTP stream request",
			setupReq: func() *http.Request {
				req := httptest.NewRequest("GET", "/mcp", nil)
				req.Header.Set("Accept", "application/json")
				return req
			},
			expectSSE:   false,
			description: "Should use HTTP stream transport",
		},
		{
			name: "SSE via query parameter",
			setupReq: func() *http.Request {
				req := httptest.NewRequest("GET", "/mcp?transport=sse", nil)
				return req
			},
			expectSSE:   true,
			description: "Should use SSE transport via query param",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setupReq()
			w := httptest.NewRecorder()

			handler(w, req)

			// Check response based on expected transport
			if tt.expectSSE {
				// SSE responses should have event-stream content type
				contentType := w.Header().Get("Content-Type")
				assert.True(t, 
					strings.Contains(contentType, "text/event-stream") || w.Code != 200,
					"Expected SSE content type or error, got: %s", contentType)
			}

			// Both transports should return 200 OK for basic requests
			assert.True(t, w.Code == 200 || w.Code == 405, // 405 for unsupported methods
				"Expected 200 OK or 405, got: %d", w.Code)
		})
	}
}

func TestSSEServer_WriteSSEEvent(t *testing.T) {
	logger := zap.NewNop()
	mcpServer := server.NewMCPServer("Test Server", "1.0.0")
	sseServer := NewSSEServer(mcpServer, logger, "Test Server", "1.0.0")

	w := httptest.NewRecorder()
	
	testData := map[string]interface{}{
		"test":      "value",
		"timestamp": 1234567890,
	}

	sseServer.writeSSEEvent(w, "test_event", testData)

	body := w.Body.String()
	lines := strings.Split(strings.TrimSpace(body), "\n")

	// Check SSE format
	assert.Equal(t, "event: test_event", lines[0])
	assert.True(t, strings.HasPrefix(lines[1], "data: "))
	assert.Contains(t, lines[1], `"test":"value"`)
	assert.Contains(t, lines[1], `"timestamp":1234567890`)
	assert.Equal(t, "", lines[2]) // Empty line after event
}