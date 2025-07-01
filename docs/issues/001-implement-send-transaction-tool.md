---
title: 'Implement send_transaction MCP Tool Business Logic'
labels: ['enhancement', 'MCP', 'native-host', 'high-priority']
assignees: []
---

## Summary

Implement the business logic for the `send_transaction` MCP tool in the Native Host to enable AI Agents to send blockchain transactions.

## Background

The `send_transaction` tool interface skeleton exists but lacks complete business logic implementation. This is a critical feature for AI-controlled trading operations.

## Requirements

### Functional Requirements

- [ ] Implement transaction creation and signing for Ethereum chains
- [ ] Support ERC-20 token transfers
- [ ] Implement gas estimation and fee calculation
- [ ] Add transaction validation and security checks
- [ ] Support transaction confirmation tracking
- [ ] Implement proper error handling for various failure scenarios

### Technical Requirements

- [ ] Update `native/pkg/mcp/tools/send_transaction_tool.go`
- [ ] Integrate with existing wallet manager and chain interfaces
- [ ] Support multiple chain types (ETH, BSC initially)
- [ ] Implement transaction status tracking
- [ ] Add comprehensive input validation

### Security Requirements

- [ ] Validate recipient addresses
- [ ] Check sufficient balance before sending
- [ ] Implement transaction amount limits
- [ ] Add confirmation mechanisms for large transactions
- [ ] Validate gas limits and prevent excessive fees

## Acceptance Criteria

- [ ] Tool can successfully send ETH transactions on Ethereum
- [ ] Tool can send ERC-20 token transactions
- [ ] Proper error messages for insufficient balance, invalid addresses, etc.
- [ ] Gas estimation works correctly
- [ ] Integration tests pass
- [ ] Security validations prevent malicious transactions

## Implementation Details

### Files to Modify

- `native/pkg/mcp/tools/send_transaction_tool.go`
- `native/pkg/wallet/transaction.go`
- `native/pkg/wallet/chain/eth_chain.go`

### API Schema

```json
{
  "name": "send_transaction",
  "description": "Send a blockchain transaction",
  "inputSchema": {
    "type": "object",
    "properties": {
      "chain": { "type": "string", "enum": ["ethereum", "bsc"] },
      "to": { "type": "string", "description": "Recipient address" },
      "amount": { "type": "string", "description": "Amount to send" },
      "token": {
        "type": "string",
        "description": "Token contract address (optional, ETH if not provided)"
      },
      "gas_limit": { "type": "number", "description": "Gas limit (optional)" },
      "gas_price": { "type": "string", "description": "Gas price in gwei (optional)" }
    },
    "required": ["chain", "to", "amount"]
  }
}
```

## Dependencies

- Requires existing wallet manager and chain interfaces
- Depends on go-ethereum library for transaction signing
- Related to issue #002 (transaction confirmation)

## Testing Requirements

- [ ] Unit tests for transaction creation and validation
- [ ] Integration tests with real blockchain interaction
- [ ] Error case testing (insufficient balance, invalid addresses)
- [ ] Gas estimation accuracy tests

## References

- Technical Spec: `docs/teck_spec.md`
- MCP API Documentation: `docs/apis/native_host_mcp_api.md`
- Existing Implementation: `native/pkg/mcp/tools/create_wallet_tool.go`
