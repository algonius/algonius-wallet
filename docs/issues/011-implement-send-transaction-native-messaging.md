---
title: 'Implement send_transaction Native Messaging RPC Method'
labels: ['enhancement', 'native-messaging', 'browser-extension', 'high-priority', 'security']
assignees: []
---

## Summary

Implement the `send_transaction` RPC method for Native Messaging to enable browser extensions to send blockchain transactions with user approval.

## Background

Browser extensions need the ability to send transactions initiated by users through the UI. This requires strong authentication and user confirmation mechanisms.

## Requirements

### Functional Requirements

- [ ] Send ETH and ERC-20 token transactions
- [ ] Support gas estimation and fee calculation
- [ ] Implement transaction validation and security checks
- [ ] Require user confirmation before sending
- [ ] Return transaction hash and status

### Technical Requirements

- [ ] Add RPC method handler in `native/pkg/messaging/native.go`
- [ ] Integrate with existing transaction and wallet services
- [ ] Implement user confirmation UI
- [ ] Add transaction queue management
- [ ] Support transaction status tracking

### Security Requirements

- [ ] Require password authentication
- [ ] Implement transaction amount limits
- [ ] Add recipient address validation
- [ ] Show transaction details for user confirmation
- [ ] Implement rate limiting for transactions

## Acceptance Criteria

- [ ] RPC method sends transactions successfully
- [ ] User confirmation is required for all transactions
- [ ] Proper error handling for various failure scenarios
- [ ] Transaction details are clearly displayed
- [ ] Gas estimation works correctly
- [ ] Integration tests pass with real transactions

## Implementation Details

### RPC Method Schema

```json
{
  "method": "send_transaction",
  "params": {
    "from": "string (required)",
    "to": "string (required)",
    "amount": "string (required)",
    "chain": "string (required, enum: [\"ethereum\", \"bsc\"])",
    "token": "string (optional, contract address)",
    "gas_limit": "number (optional)",
    "gas_price": "string (optional)",
    "data": "string (optional, hex data)",
    "password": "string (required)"
  },
  "result": {
    "transaction_hash": "string",
    "status": "string",
    "gas_used": "number",
    "block_number": "number (optional)",
    "sent_at": "number (timestamp)"
  },
  "error": {
    "code": "number",
    "message": "string"
  }
}
```

### Files to Modify

- `native/pkg/messaging/native.go` - Add RPC method registration
- `native/pkg/wallet/transaction.go` - Add transaction sending logic
- `native/pkg/ui/confirmation.go` - Add user confirmation UI
- `native/pkg/wallet/validation.go` - Add transaction validation

### Error Codes

- `-32031`: Invalid password
- `-32032`: Insufficient balance
- `-32033`: Invalid recipient address
- `-32034`: Transaction rejected by user
- `-32035`: Gas estimation failed
- `-32036`: Transaction broadcast failed

## Dependencies

- Requires user confirmation UI system
- Depends on existing wallet and transaction services
- Related to gas estimation and fee calculation

## Testing Requirements

- [ ] Unit tests for transaction validation
- [ ] Integration tests with user confirmation
- [ ] Security tests for authentication
- [ ] Error case testing (insufficient balance, invalid addresses)
- [ ] Gas estimation accuracy tests

## References

- Technical Spec: `docs/teck_spec.md`
- Native Messaging: `native/pkg/messaging/native.go`
- Related MCP Tool: Issue #001 (send_transaction MCP tool)
