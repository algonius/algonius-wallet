---
title: 'Implement get_wallet_info Native Messaging RPC Method'
labels: ['enhancement', 'native-messaging', 'browser-extension', 'medium-priority']
assignees: []
---

## Summary

Implement the `get_wallet_info` RPC method for Native Messaging to enable browser extensions to retrieve wallet information including balances, addresses, and status.

## Background

Browser extensions need to display wallet information to users, including addresses, balances, and connection status. This information should be available without requiring high-level authentication.

## Requirements

### Functional Requirements

- [ ] Return wallet addresses for all supported chains
- [ ] Include current balances for each address
- [ ] Show wallet connection status
- [ ] Support filtering by chain or address
- [ ] Return wallet metadata (creation date, alias, etc.)

### Technical Requirements

- [ ] Add RPC method handler in `native/pkg/messaging/native.go`
- [ ] Integrate with existing balance and wallet services
- [ ] Implement efficient data caching
- [ ] Support pagination for large wallet collections
- [ ] Add real-time balance updates

### Security Requirements

- [ ] Return only public wallet information
- [ ] Never expose private keys or mnemonics
- [ ] Implement basic rate limiting
- [ ] Validate request parameters

## Acceptance Criteria

- [ ] RPC method returns complete wallet information
- [ ] Balances are accurate and up-to-date
- [ ] Proper error handling for invalid requests
- [ ] Response includes all public wallet data
- [ ] Performance is acceptable for multiple wallets
- [ ] Integration tests pass with various wallet states

## Implementation Details

### RPC Method Schema

```json
{
  "method": "get_wallet_info",
  "params": {
    "chain": "string (optional, enum: [\"ethereum\", \"bsc\", \"all\"])",
    "address": "string (optional)",
    "include_balances": "boolean (optional, default: true)",
    "include_tokens": "boolean (optional, default: false)"
  },
  "result": {
    "wallets": [
      {
        "address": "string",
        "chain": "string",
        "balance": "string",
        "tokens": [
          {
            "contract": "string",
            "symbol": "string",
            "balance": "string",
            "decimals": "number"
          }
        ],
        "created_at": "number (timestamp)",
        "alias": "string",
        "is_active": "boolean"
      }
    ],
    "total_count": "number",
    "last_updated": "number (timestamp)"
  },
  "error": {
    "code": "number",
    "message": "string"
  }
}
```

### Files to Modify

- `native/pkg/messaging/native.go` - Add RPC method registration
- `native/pkg/wallet/manager.go` - Add info retrieval functionality
- `native/pkg/wallet/info.go` - Add wallet info structures
- `native/pkg/cache/balance.go` - Add balance caching

### Error Codes

- `-32021`: Invalid chain parameter
- `-32022`: Invalid address format
- `-32023`: Wallet not found
- `-32024`: Balance retrieval failed

## Dependencies

- Requires balance retrieval services
- Depends on wallet manager implementation
- Related to existing chain interfaces

## Testing Requirements

- [ ] Unit tests for info retrieval
- [ ] Integration tests with multiple wallets
- [ ] Performance tests for large wallet collections
- [ ] Error case testing (invalid chains, addresses)
- [ ] Balance accuracy tests

## References

- Technical Spec: `docs/teck_spec.md`
- Native Messaging: `native/pkg/messaging/native.go`
- Related MCP Tool: `get_balance_tool.go`
