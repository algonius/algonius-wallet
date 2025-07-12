# SSE (Server-Sent Events) Implementation Guide

## Overview

The Algonius Native Host now supports Server-Sent Events (SSE) for real-time event streaming. This implementation uses the `mcp-go` library's built-in SSE server functionality to provide live updates for wallet operations, transactions, and status changes.

## Implementation Details

### Architecture

The SSE implementation consists of several key components:

1. **Event Broadcasting System** (`pkg/events/broadcaster.go`)
   - Central event management and distribution
   - Session-based event broadcasting
   - Thread-safe operations with mutex protection

2. **SSE Events Resource** (`pkg/mcp/resources/sse_events_resource.go`)
   - MCP resource for SSE event information
   - Provides event type documentation and usage examples

3. **SSE Server Integration** (`cmd/main.go`)
   - Uses `server.NewSSEServer()` from mcp-go library
   - Configurable via `MCP_SERVER_TYPE=sse` environment variable

### Event Types

The system supports the following event types:

- `transaction_confirmed` - Transaction successfully confirmed on blockchain
- `transaction_pending` - Transaction submitted and waiting for confirmation
- `transaction_failed` - Transaction failed to execute
- `balance_changed` - Wallet balance updated
- `wallet_status_changed` - Wallet connection status changed
- `block_new` - New block detected on monitored chains

### Usage

#### Starting the SSE Server

```bash
# Set environment variable to enable SSE mode
export MCP_SERVER_TYPE=sse

# Start the native host
./bin/algonius-wallet-host
```

The server will start on the default port (configurable) and provide SSE endpoints.

#### Connecting to SSE Events

```bash
# Connect with curl for testing
curl -N -H 'Accept: text/event-stream' http://localhost:9444/sse/events
```

#### Event Format

Events are sent in standard SSE format with JSON payloads:

```
event: transaction_confirmed
data: {
  "id": "evt_123456789",
  "type": "transaction_confirmed",
  "timestamp": "2024-01-15T10:30:00Z",
  "chain": "ethereum",
  "data": {
    "tx_hash": "0x1234567890abcdef...",
    "from": "0xfrom...",
    "to": "0xto...",
    "amount": "1.5",
    "token": "ETH",
    "gas_used": "21000",
    "block": "12345"
  }
}
```

### Integration with MCP Tools

SSE events are automatically generated when using MCP tools:

1. **Send Transaction Tool** - Generates `transaction_pending` events
2. **Confirm Transaction Tool** - Generates `transaction_confirmed` or `transaction_failed` events
3. **Balance Queries** - May trigger `balance_changed` events
4. **Wallet Operations** - Generate `wallet_status_changed` events

### Client Implementation

#### JavaScript/TypeScript Example

```typescript
const eventSource = new EventSource('http://localhost:9444/sse/events');

eventSource.addEventListener('transaction_confirmed', (event) => {
  const data = JSON.parse(event.data);
  console.log('Transaction confirmed:', data);
});

eventSource.addEventListener('balance_changed', (event) => {
  const data = JSON.parse(event.data);
  console.log('Balance updated:', data);
});

eventSource.onerror = (error) => {
  console.error('SSE connection error:', error);
};
```

#### React Hook Example

```typescript
import { useEffect, useState } from 'react';

interface WalletEvent {
  id: string;
  type: string;
  timestamp: string;
  chain?: string;
  data: any;
}

export function useWalletEvents(serverUrl: string) {
  const [events, setEvents] = useState<WalletEvent[]>([]);
  const [connectionStatus, setConnectionStatus] = useState<'connecting' | 'connected' | 'disconnected'>('disconnected');

  useEffect(() => {
    const eventSource = new EventSource(`${serverUrl}/sse/events`);
    
    setConnectionStatus('connecting');

    eventSource.onopen = () => {
      setConnectionStatus('connected');
    };

    eventSource.onmessage = (event) => {
      try {
        const walletEvent: WalletEvent = JSON.parse(event.data);
        setEvents(prev => [...prev, walletEvent]);
      } catch (error) {
        console.error('Failed to parse event:', error);
      }
    };

    eventSource.onerror = () => {
      setConnectionStatus('disconnected');
    };

    return () => {
      eventSource.close();
    };
  }, [serverUrl]);

  return { events, connectionStatus };
}
```

### Testing

#### Unit Tests

Run the SSE integration tests:

```bash
cd native
INTEGRATION_TESTS=true go test -v ./tests/integration/sse_events_test.go
```

#### Manual Testing

1. Start the SSE server:
```bash
cd native
MCP_SERVER_TYPE=sse ./bin/algonius-wallet-host
```

2. Connect with curl in another terminal:
```bash
curl -N -H 'Accept: text/event-stream' http://localhost:9444/sse/events
```

3. Trigger wallet operations to see events in real-time

### Performance Considerations

- **Connection Limits**: The server can handle multiple concurrent SSE connections
- **Memory Management**: Event history is not stored; only real-time events are broadcast
- **Resource Usage**: Each active SSE connection consumes minimal resources
- **Scalability**: The event broadcaster is designed for concurrent operations

### Error Handling

The SSE implementation includes robust error handling:

- **Connection Failures**: Clients should implement reconnection logic
- **Event Processing Errors**: Invalid events are logged but don't affect other connections
- **Server Shutdown**: All active connections are gracefully closed

### Security Considerations

- **CORS**: Configure appropriate CORS headers for browser-based clients
- **Authentication**: Consider implementing authentication for production use
- **Rate Limiting**: Monitor and limit connection rates to prevent abuse

### Troubleshooting

#### Common Issues

1. **Connection Refused**
   - Verify the server is running with `MCP_SERVER_TYPE=sse`
   - Check the correct port is being used

2. **No Events Received**
   - Ensure wallet operations are being performed
   - Check event broadcaster is properly initialized

3. **Connection Drops**
   - Implement client-side reconnection logic
   - Check network stability

#### Debug Mode

Enable debug logging:

```bash
export LOG_LEVEL=debug
MCP_SERVER_TYPE=sse ./bin/algonius-wallet-host
```

### Future Enhancements

Planned improvements include:

1. **Event Filtering**: Allow clients to subscribe to specific event types
2. **Event History**: Optional event replay functionality
3. **Authentication**: Secure SSE endpoints with authentication
4. **Metrics**: Performance monitoring and connection metrics
5. **Event Aggregation**: Batch events for high-frequency scenarios

## Conclusion

The SSE implementation provides a robust, real-time event streaming solution for the Algonius Wallet. It leverages the proven `mcp-go` library and follows industry standards for Server-Sent Events, ensuring compatibility with modern web browsers and client applications.
