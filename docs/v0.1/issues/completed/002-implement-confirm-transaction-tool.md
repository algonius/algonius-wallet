---
title: 'Implement confirm_transaction MCP Tool'
labels: ['enhancement', 'MCP', 'native-host', 'high-priority', 'completed']
assignees: []
---

## Summary

Implement the `confirm_transaction` MCP tool to allow AI Agents to check transaction status and confirmations on the blockchain.

## Background

Transaction confirmation is critical for AI trading operations. The tool needs to track transaction status, confirmation count, and provide real-time updates on transaction progress.

## Requirements

### Functional Requirements

- [x] Check transaction status by transaction hash
- [x] Return confirmation count and required confirmations
- [x] Support multiple blockchain networks (ETH, BSC)
- [x] Provide detailed transaction information (block number, gas used, fee)
- [x] Handle unconfirmed and failed transactions
- [x] Support custom confirmation thresholds

### Technical Requirements

- [x] Create `native/pkg/mcp/tools/confirm_transaction_tool.go`
- [x] Integrate with blockchain explorers APIs
- [x] Implement transaction status parsing
- [x] Add chain-specific confirmation logic
- [x] Use structured response format
- [x] Handle network errors gracefully

### Security Requirements

- [x] Validate transaction hash format
- [x] Prevent information leakage through invalid requests
- [x] Implement rate limiting
- [x] Secure API key handling for blockchain explorers

## Acceptance Criteria

- [x] Tool can successfully check transaction status on Ethereum
- [x] Tool can check transaction status on BSC
- [x] Proper error handling for invalid transaction hashes
- [x] Response includes all necessary confirmation details
- [x] Integration tests pass with real blockchain data
- [x] Security validations prevent malicious requests

## Implementation Details

### Files Created/Modified

- `native/pkg/mcp/tools/confirm_transaction_tool.go` (new)
- `native/pkg/mcp/tools/confirm_transaction_tool_test.go` (new)
- `native/pkg/wallet/transaction.go` (extended with ConfirmTransaction function)
- `native/pkg/wallet/chain/eth_chain.go` (extended with transaction confirmation)
- `native/pkg/wallet/chain/bsc_chain.go` (extended with transaction confirmation)
- `native/cmd/main.go` (registered new tool)

### API Schema

```json
{
  "name": "confirm_transaction",
  "description": "Check transaction confirmation status",
  "input_schema": {
    "type": "object",
    "properties": {
      "chain": {
        "type": "string",
        "description": "Chain identifier (ethereum, bsc)"
      },
      "tx_hash": {
        "type": "string",
        "description": "Transaction hash"
      },
      "required_confirmations": {
        "type": "number",
        "description": "Required confirmations for finality (default: 6 for Ethereum, 3 for BSC)"
      }
    },
    "required": ["chain", "tx_hash"]
  },
  "output_schema": {
    "type": "object",
    "properties": {
      "tx_hash": { "type": "string" },
      "chain": { "type": "string" },
      "status": { "type": "string", "enum": ["pending", "confirmed", "failed"] },
      "confirmations": { "type": "integer" },
      "required_confirmations": { "type": "integer" },
      "block_number": { "type": "integer" },
      "gas_used": { "type": "string" },
      "transaction_fee": { "type": "string" },
      "timestamp": { "type": "string", "format": "date-time" }
    },
    "required": ["tx_hash", "chain", "status", "confirmations", "required_confirmations"]
  }
}
```

### Response Format

The tool returns a markdown-formatted response with:

1. Transaction hash
2. Chain information
3. Status (pending, confirmed, or failed)
4. Confirmation count and required confirmations
5. Block number
6. Gas used
7. Transaction fee
8. Timestamp

### Key Features Implemented

1. **Multi-chain Support**: Works with Ethereum and BSC
2. **Detailed Information**: Provides comprehensive transaction details
3. **Flexible Confirmations**: Supports custom confirmation thresholds
4. **Error Handling**: Handles network errors and invalid requests gracefully
5. **Security Validations**: Validates inputs to prevent malicious requests

## Dependencies

- Depends on blockchain RPC providers (Infura, Alchemy, etc.)
- Requires go-ethereum library for Ethereum chains
- Related to issue #001 (send_transaction tool)

## Testing Requirements

- [x] Unit tests for transaction confirmation logic
- [x] Integration tests with real blockchain data
- [x] Test various chain types (ETH, BSC)
- [x] Test error cases (invalid hash, network errors)
- [x] Test different transaction statuses (pending, confirmed, failed)
- [x] Security testing for malicious inputs

## Configuration

- [ ] RPC endpoint configuration per chain
- [ ] Default confirmation thresholds
- [ ] Cache TTL settings
- [ ] Rate limiting parameters

## References

- Technical Spec: `docs/teck_spec.md`
- MCP API Documentation: `docs/apis/native_host_mcp_api.md`
- Ethereum JSON-RPC API: https://ethereum.org/en/developers/docs/apis/json-rpc/
