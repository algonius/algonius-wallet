# MCP SSE Transport Layer Support

## Overview

The Algonius Wallet Native Host now supports dual MCP transport layers:
- **HTTP Stream Transport** (original implementation)
- **SSE (Server-Sent Events) Transport** (new implementation)

This dual transport support ensures compatibility with a wider range of MCP clients, including tools like Cline that require SSE transport layer.

## Transport Auto-Detection

The server automatically detects the preferred transport method based on the client request:

### SSE Transport Detection
The server will use SSE transport when:
- `Accept: text/event-stream` header is present
- `transport=sse` query parameter is provided
- User-Agent contains "eventsource"

### HTTP Stream Transport (Default)
All other requests will use the original HTTP stream transport.

## Endpoints

The server provides multiple endpoints for flexibility:

- **`/mcp`** - Main endpoint with automatic transport detection
- **`/mcp/sse`** - Explicit SSE transport endpoint
- **`/mcp/stream`** - Explicit HTTP stream transport endpoint

## Usage Examples

### For SSE-Compatible Clients (like Cline)

```javascript
// Option 1: Using Accept header
const response = await fetch('http://localhost:9444/mcp', {
  headers: {
    'Accept': 'text/event-stream'
  }
});

// Option 2: Using explicit SSE endpoint
const response = await fetch('http://localhost:9444/mcp/sse');

// Option 3: Using query parameter
const response = await fetch('http://localhost:9444/mcp?transport=sse');

// Option 4: Using EventSource for streaming
const eventSource = new EventSource('http://localhost:9444/mcp/sse');
eventSource.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Received:', data);
};
```

### For HTTP Stream Clients (Original)

```javascript
// Standard HTTP request (default behavior)
const response = await fetch('http://localhost:9444/mcp', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    jsonrpc: '2.0',
    method: 'tools/list',
    id: 1
  })
});

// Or use explicit stream endpoint
const response = await fetch('http://localhost:9444/mcp/stream', {
  // ... same configuration
});
```

## SSE Event Format

SSE events follow the standard Server-Sent Events format with JSON-RPC 2.0 content:

```
event: connected
data: {"timestamp": 1640995200, "status": "ready"}

event: server_info
data: {"name": "Algonius Native Host", "version": "0.1.0", "capabilities": {...}}

event: response
data: {"jsonrpc": "2.0", "id": 1, "result": {"tools": [...]}}

event: heartbeat
data: {"timestamp": 1640995230}

event: error
data: {"error": "parse_error", "message": "Invalid JSON-RPC request"}
```

## Supported MCP Methods

Both transport layers support all standard MCP methods:

- `initialize` - Initialize MCP connection
- `tools/list` - List available tools
- `tools/call` - Call a specific tool
- `resources/list` - List available resources
- `resources/read` - Read a specific resource

## Configuration

The server port can be configured via environment variable:

```bash
export SSE_PORT=":9444"  # Default port
./algonius-wallet-host
```

## Client Compatibility

### Compatible with SSE Transport
- ✅ Cline
- ✅ Standard EventSource API
- ✅ Any client that sends `Accept: text/event-stream`

### Compatible with HTTP Stream Transport
- ✅ Original mcp-go clients
- ✅ Custom HTTP clients
- ✅ All existing integrations

## Implementation Details

### Transport Detection Logic

```go
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
```

### Dual Handler Implementation

```go
func CreateMCPHandler(mcpServer *server.MCPServer, logger *zap.Logger) http.HandlerFunc {
    sseServer := NewSSEServer(mcpServer, logger)
    streamServer := server.NewStreamableHTTPServer(mcpServer)

    return func(w http.ResponseWriter, r *http.Request) {
        if SSETransportDetector(r) {
            sseServer.ServeHTTP(w, r)
        } else {
            streamServer.ServeHTTP(w, r)
        }
    }
}
```

## Testing

Run the test suite to verify both transport layers work correctly:

```bash
cd native
go test ./pkg/mcp -v
```

The tests cover:
- Transport detection logic
- SSE event formatting
- JSON-RPC message handling
- Multi-transport handler routing
- Error handling for both transports

## Troubleshooting

### Client Not Using SSE Transport

1. **Check Accept Header**: Ensure client sends `Accept: text/event-stream`
2. **Use Explicit Endpoint**: Try `/mcp/sse` endpoint directly
3. **Add Query Parameter**: Use `?transport=sse` in URL
4. **Check Logs**: Server logs will indicate which transport is being used

### SSE Connection Issues

1. **CORS Headers**: SSE includes CORS headers for cross-origin requests
2. **Connection Keep-Alive**: Server sends heartbeat events every 30 seconds
3. **Reconnection**: Client should implement reconnection logic for dropped connections

### Performance Considerations

- **Concurrent Connections**: Both transports support multiple concurrent clients
- **Memory Usage**: SSE maintains connections longer; monitor memory usage
- **Heartbeat Frequency**: Adjust heartbeat interval if needed for your use case

## Migration Guide

### For Existing HTTP Stream Clients
No changes required - existing clients will continue to work unchanged.

### For New SSE Clients
Simply use one of the SSE detection methods described above.

### For Libraries Adding SSE Support
Implement one of the transport detection patterns:
- Send `Accept: text/event-stream` header
- Use `/mcp/sse` endpoint
- Add `transport=sse` query parameter