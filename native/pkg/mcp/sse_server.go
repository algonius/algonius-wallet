// SPDX-License-Identifier: Apache-2.0
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// SSEServer implements MCP protocol over Server-Sent Events
type SSEServer struct {
	mcpServer   *server.MCPServer
	logger      *zap.Logger
	serverName  string
	serverVersion string
}

// NewSSEServer creates a new SSE-based MCP server
func NewSSEServer(mcpServer *server.MCPServer, logger *zap.Logger, serverName, serverVersion string) *SSEServer {
	return &SSEServer{
		mcpServer:     mcpServer,
		logger:        logger,
		serverName:    serverName,
		serverVersion: serverVersion,
	}
}

// ServeHTTP implements the http.Handler interface for SSE transport
func (s *SSEServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Only support GET and POST for SSE
	if r.Method != "GET" && r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.logger.Info("SSE client connected", zap.String("remote_addr", r.RemoteAddr))

	// Initialize the connection by sending a connection event
	s.writeSSEEvent(w, "connected", map[string]interface{}{
		"timestamp": time.Now().Unix(),
		"status":    "ready",
	})

	// Handle the connection based on method
	if r.Method == "GET" {
		s.handleSSEStream(w, r)
	} else if r.Method == "POST" {
		s.handleSSERequest(w, r)
	}
}

// handleSSEStream handles GET requests for persistent SSE connections
func (s *SSEServer) handleSSEStream(w http.ResponseWriter, r *http.Request) {
	// Create a flusher for real-time streaming
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Send initial server info
	serverInfo := map[string]interface{}{
		"name":         s.serverName,
		"version":      s.serverVersion,
		"capabilities": s.getServerCapabilities(),
	}
	s.writeSSEEvent(w, "server_info", serverInfo)
	flusher.Flush()

	// Keep connection alive with heartbeats
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Listen for client disconnect
	ctx := r.Context()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("SSE client disconnected", zap.String("remote_addr", r.RemoteAddr))
			return
		case <-ticker.C:
			s.writeSSEEvent(w, "heartbeat", map[string]interface{}{
				"timestamp": time.Now().Unix(),
			})
			flusher.Flush()
		}
	}
}

// handleSSERequest handles POST requests for single MCP calls over SSE
func (s *SSEServer) handleSSERequest(w http.ResponseWriter, r *http.Request) {
	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.logger.Error("Failed to read request body", zap.Error(err))
		s.writeSSEError(w, "invalid_request", "Failed to read request body")
		return
	}

	// Parse JSON-RPC request
	var rpcRequest mcp.JSONRPCRequest
	if err := json.Unmarshal(body, &rpcRequest); err != nil {
		s.logger.Error("Failed to parse JSON-RPC request", zap.Error(err))
		s.writeSSEError(w, "parse_error", "Invalid JSON-RPC request")
		return
	}

	s.logger.Info("Processing MCP request over SSE", 
		zap.String("method", rpcRequest.Method), 
		zap.Any("id", rpcRequest.ID))

	// Process the MCP request
	response := s.processMCPRequest(&rpcRequest)

	// Send response as SSE event
	s.writeSSEEvent(w, "response", response)

	// Flush the response
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}

// processMCPRequest processes a single MCP JSON-RPC request
func (s *SSEServer) processMCPRequest(request *mcp.JSONRPCRequest) interface{} {
	switch request.Method {
	case "initialize":
		return s.handleInitialize(request)
	case "tools/list":
		return s.handleToolsList(request)
	case "tools/call":
		return s.handleToolsCall(request)
	case "resources/list":
		return s.handleResourcesList(request)
	case "resources/read":
		return s.handleResourcesRead(request)
	default:
		return mcp.JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error: &mcp.JSONRPCError{
				Code:    -32601,
				Message: "Method not found",
			},
		}
	}
}

// handleInitialize handles the MCP initialize request
func (s *SSEServer) handleInitialize(request *mcp.JSONRPCRequest) interface{} {
	return mcp.JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    s.getServerCapabilities(),
			"serverInfo": map[string]interface{}{
				"name":    s.serverName,
				"version": s.serverVersion,
			},
		},
	}
}

// handleToolsList handles tools/list requests
func (s *SSEServer) handleToolsList(request *mcp.JSONRPCRequest) interface{} {
	ctx := context.Background()
	result, err := s.mcpServer.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return mcp.JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error: &mcp.JSONRPCError{
				Code:    -32000,
				Message: err.Error(),
			},
		}
	}
	return mcp.JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result:  result,
	}
}

// handleToolsCall handles tools/call requests
func (s *SSEServer) handleToolsCall(request *mcp.JSONRPCRequest) interface{} {
	// Parse tool call parameters
	var params struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}

	if request.Params != nil {
		// Convert any to JSON bytes then unmarshal to our struct
		paramsBytes, err := json.Marshal(request.Params)
		if err != nil {
			return mcp.JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      request.ID,
				Error: &mcp.JSONRPCError{
					Code:    -32602,
					Message: "Invalid params",
				},
			}
		}
		if err := json.Unmarshal(paramsBytes, &params); err != nil {
			return mcp.JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      request.ID,
				Error: &mcp.JSONRPCError{
					Code:    -32602,
					Message: "Invalid params",
				},
			}
		}
	}

	// Call the tool using correct API
	ctx := context.Background()
	callReq := mcp.CallToolRequest{
		Name:      params.Name,
		Arguments: params.Arguments,
	}
	result, err := s.mcpServer.CallTool(ctx, callReq)
	if err != nil {
		return mcp.JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error: &mcp.JSONRPCError{
				Code:    -32000,
				Message: err.Error(),
			},
		}
	}

	return mcp.JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result:  result,
	}
}

// handleResourcesList handles resources/list requests
func (s *SSEServer) handleResourcesList(request *mcp.JSONRPCRequest) interface{} {
	ctx := context.Background()
	result, err := s.mcpServer.ListResources(ctx, mcp.ListResourcesRequest{})
	if err != nil {
		return mcp.JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error: &mcp.JSONRPCError{
				Code:    -32000,
				Message: err.Error(),
			},
		}
	}
	return mcp.JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result:  result,
	}
}

// handleResourcesRead handles resources/read requests
func (s *SSEServer) handleResourcesRead(request *mcp.JSONRPCRequest) interface{} {
	// Parse resource read parameters
	var params struct {
		URI string `json:"uri"`
	}

	if request.Params != nil {
		// Convert any to JSON bytes then unmarshal to our struct
		paramsBytes, err := json.Marshal(request.Params)
		if err != nil {
			return mcp.JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      request.ID,
				Error: &mcp.JSONRPCError{
					Code:    -32602,
					Message: "Invalid params",
				},
			}
		}
		if err := json.Unmarshal(paramsBytes, &params); err != nil {
			return mcp.JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      request.ID,
				Error: &mcp.JSONRPCError{
					Code:    -32602,
					Message: "Invalid params",
				},
			}
		}
	}

	// Read the resource using correct API
	ctx := context.Background()
	readReq := mcp.ReadResourceRequest{
		URI: params.URI,
	}
	result, err := s.mcpServer.ReadResource(ctx, readReq)
	if err != nil {
		return mcp.JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error: &mcp.JSONRPCError{
				Code:    -32000,
				Message: err.Error(),
			},
		}
	}

	return mcp.JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result:  result,
	}
}

// getServerCapabilities returns the server capabilities
func (s *SSEServer) getServerCapabilities() map[string]interface{} {
	return map[string]interface{}{
		"tools": map[string]interface{}{
			"listChanged": false,
		},
		"resources": map[string]interface{}{
			"subscribe":   false,
			"listChanged": false,
		},
		"logging": map[string]interface{}{},
	}
}

// writeSSEEvent writes a Server-Sent Event to the response writer
func (s *SSEServer) writeSSEEvent(w http.ResponseWriter, eventType string, data interface{}) {
	// Convert data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		s.logger.Error("Failed to marshal SSE event data", zap.Error(err))
		return
	}

	// Write SSE event format
	fmt.Fprintf(w, "event: %s\n", eventType)
	fmt.Fprintf(w, "data: %s\n\n", string(jsonData))
}

// writeSSEError writes an error event to the response writer
func (s *SSEServer) writeSSEError(w http.ResponseWriter, errorType string, message string) {
	errorData := map[string]interface{}{
		"error":   errorType,
		"message": message,
	}
	s.writeSSEEvent(w, "error", errorData)
}

// SSETransportDetector detects if a request wants SSE transport
func SSETransportDetector(r *http.Request) bool {
	// Check Accept header for text/event-stream
	accept := r.Header.Get("Accept")
	if strings.Contains(accept, "text/event-stream") {
		return true
	}

	// Check for SSE-specific query parameters
	if r.URL.Query().Get("transport") == "sse" {
		return true
	}

	// Check for EventSource user agent patterns
	userAgent := r.Header.Get("User-Agent")
	if strings.Contains(strings.ToLower(userAgent), "eventsource") {
		return true
	}

	return false
}

// CreateMCPHandler creates a handler that supports both HTTP stream and SSE transports
func CreateMCPHandler(mcpServer *server.MCPServer, logger *zap.Logger, serverName, serverVersion string) http.HandlerFunc {
	// Create SSE server
	sseServer := NewSSEServer(mcpServer, logger, serverName, serverVersion)
	
	// Create HTTP stream server
	streamServer := server.NewStreamableHTTPServer(mcpServer)

	return func(w http.ResponseWriter, r *http.Request) {
		// Detect transport preference
		if SSETransportDetector(r) {
			logger.Info("Using SSE transport", zap.String("remote_addr", r.RemoteAddr))
			sseServer.ServeHTTP(w, r)
		} else {
			logger.Info("Using HTTP stream transport", zap.String("remote_addr", r.RemoteAddr))
			streamServer.ServeHTTP(w, r)
		}
	}
}