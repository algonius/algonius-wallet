---
title: 'Implement send_transaction MCP Tool Business Logic'
labels: ['enhancement', 'MCP', 'native-host', 'high-priority', 'completed']
assignees: []
---

## Summary

Implement the business logic for the `send_transaction` MCP tool in the Native Host to enable AI Agents to send blockchain transactions.

## Background

The `send_transaction` tool interface skeleton exists but lacks complete business logic implementation. This is a critical feature for AI-controlled trading operations.

## Requirements

### Functional Requirements

- [x] Implement transaction creation and signing for Ethereum chains
- [x] Support ERC-20 token transfers
- [x] Implement gas estimation and fee calculation
- [x] Add transaction validation and security checks
- [x] Support transaction confirmation tracking
- [x] Implement proper error handling for various failure scenarios

### Technical Requirements

- [x] Update `native/pkg/mcp/tools/send_transaction_tool.go`
- [x] Integrate with existing wallet manager and chain interfaces
- [x] Support multiple chain types (ETH, BSC initially)
- [x] Implement transaction status tracking
- [x] Add comprehensive input validation

### Security Requirements

- [x] Validate recipient addresses
- [x] Check sufficient balance before sending
- [x] Implement transaction amount limits
- [x] Add confirmation mechanisms for large transactions
- [x] Validate gas limits and prevent excessive fees

## Acceptance Criteria

- [x] Tool can successfully send ETH transactions on Ethereum
- [x] Tool can send ERC-20 token transactions
- [x] Proper error messages for insufficient balance, invalid addresses, etc.
- [x] Gas estimation works correctly
- [x] Integration tests pass
- [x] Security validations prevent malicious transactions

## Implementation Details

### Files Created/Modified

- `native/pkg/mcp/tools/send_transaction_tool.go` (updated with full implementation)
- `native/pkg/wallet/manager.go` (extended with SendTransaction and EstimateGas methods)
- `native/pkg/wallet/chain/*` (extended with chain-specific transaction sending)
- `native/cmd/main.go` (registered tool with MCP server)

### API Schema

```json
{
  "name": "send_transaction",
  "description": "Send a blockchain transaction",
  "input_schema": {
    "type": "object",
    "properties": {
      "chain": {
        "type": "string",
        "description": "Chain identifier (ethereum, bsc)"
      },
      "from": {
        "type": "string",
        "description": "Sender address"
      },
      "to": {
        "type": "string",
        "description": "Recipient address"
      },
      "amount": {
        "type": "string",
        "description": "Amount to send"
      },
      "token": {
        "type": "string",
        "description": "Token contract address (optional, native token if not provided)"
      },
      "gas_limit": {
        "type": "number",
        "description": "Gas limit (optional)"
      },
      "gas_price": {
        "type": "string",
        "description": "Gas price in gwei (optional)"
      }
    },
    "required": ["chain", "from", "to", "amount"]
  },
  "output_schema": {
    "type": "object",
    "properties": {
      "chain": { "type": "string" },
      "from": { "type": "string" },
      "to": { "type": "string" },
      "amount": { "type": "string" },
      "token": { "type": "string" },
      "gas_limit": { "type": "number" },
      "gas_price": { "type": "string" },
      "tx_hash": { "type": "string" },
      "status": { "type": "string", "enum": ["pending", "confirmed", "failed"] }
    },
    "required": ["tx_hash", "status"]
  }
}
```

### Response Format

The tool returns a markdown-formatted response with:

1. Chain information
2. Transaction details (from, to, amount, token)
3. Gas information (limit and price)
4. Transaction hash
5. Status (initially "pending")

### Key Features Implemented

1. **Multi-chain Support**: Works with Ethereum and BSC
2. **Token Transfers**: Supports both native tokens and ERC-20 tokens
3. **Gas Estimation**: Automatically estimates gas if not provided
4. **Comprehensive Validation**: Validates addresses, amounts, and chain support
5. **Error Handling**: Provides clear error messages for various failure scenarios
6. **Security Checks**: Validates transactions before sending

## Dependencies

- Requires existing wallet manager and chain interfaces
- Depends on go-ethereum library for transaction signing
- Related to issue #002 (transaction confirmation)

## Testing Requirements

- [x] Unit tests for transaction creation and validation
- [x] Integration tests with real blockchain interaction
- [x] Test various token types (native and ERC-20)
- [x] Test error cases (insufficient balance, invalid addresses)
- [x] Test gas estimation functionality
- [x] Security testing for malicious inputs

## References

- Technical Spec: `docs/teck_spec.md`
- MCP API Documentation: `docs/apis/native_host_mcp_api.md`
- Existing Implementation: `native/pkg/mcp/tools/create_wallet_tool.go`
