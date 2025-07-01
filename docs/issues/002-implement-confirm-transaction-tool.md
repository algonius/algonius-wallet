---
title: 'Implement confirm_transaction MCP Tool'
labels: ['enhancement', 'MCP', 'native-host', 'high-priority']
assignees: []
---

## Summary

Implement the `confirm_transaction` MCP tool to allow AI Agents to check transaction status and confirmations on the blockchain.

## Background

Transaction confirmation is critical for AI trading operations. The tool needs to track transaction status, confirmation count, and provide real-time updates on transaction progress.

## Requirements

### Functional Requirements

- [ ] Check transaction status by transaction hash
- [ ] Return confirmation count and required confirmations
- [ ] Support multiple blockchain networks (ETH, BSC, Solana)
- [ ] Provide transaction details (gas used, block number, status)
- [ ] Handle pending, confirmed, and failed transaction states
- [ ] Support batch transaction status checking

### Technical Requirements

- [ ] Create `native/pkg/mcp/tools/confirm_transaction_tool.go`
- [ ] Integrate with blockchain RPC providers
- [ ] Implement retry logic for network failures
- [ ] Add caching for recent transaction queries
- [ ] Support configurable confirmation thresholds

### Performance Requirements

- [ ] Response time under 2 seconds for cached queries
- [ ] Support concurrent transaction status checks
- [ ] Implement rate limiting for RPC calls
- [ ] Cache transaction results to reduce API calls

## Acceptance Criteria

- [ ] Tool returns accurate transaction status for valid hashes
- [ ] Proper error handling for invalid transaction hashes
- [ ] Supports all target blockchain networks
- [ ] Confirmation count matches blockchain state
- [ ] Integration tests pass with real transactions
- [ ] Performance requirements met

## Implementation Details

### Files to Create/Modify

- `native/pkg/mcp/tools/confirm_transaction_tool.go` (new)
- `native/pkg/wallet/transaction.go` (extend)
- `native/pkg/wallet/chain/interfaces.go` (add methods)
- `native/pkg/wallet/chain/eth_chain.go` (implement methods)
- `native/cmd/main.go` (register tool)

### API Schema

```json
{
  "name": "confirm_transaction",
  "description": "Check transaction confirmation status",
  "inputSchema": {
    "type": "object",
    "properties": {
      "chain": { "type": "string", "enum": ["ethereum", "bsc", "solana"] },
      "tx_hash": { "type": "string", "description": "Transaction hash" },
      "required_confirmations": {
        "type": "number",
        "default": 6,
        "description": "Required confirmations"
      }
    },
    "required": ["chain", "tx_hash"]
  }
}
```

### Response Format

```json
{
  "status": "confirmed|pending|failed",
  "confirmations": 12,
  "required_confirmations": 6,
  "block_number": 18500000,
  "gas_used": "21000",
  "transaction_fee": "0.001",
  "timestamp": "2025-06-24T07:00:00Z"
}
```

## Dependencies

- Depends on blockchain RPC providers (Infura, Alchemy, etc.)
- Requires go-ethereum library for Ethereum chains
- Related to issue #001 (send_transaction tool)

## Testing Requirements

- [ ] Unit tests for transaction status parsing
- [ ] Integration tests with real blockchain data
- [ ] Mock tests for network failure scenarios
- [ ] Performance tests for concurrent requests
- [ ] Cache effectiveness tests

## Configuration

- [ ] RPC endpoint configuration per chain
- [ ] Default confirmation thresholds
- [ ] Cache TTL settings
- [ ] Rate limiting parameters

## References

- Technical Spec: `docs/teck_spec.md`
- MCP API Documentation: `docs/apis/native_host_mcp_api.md`
- Ethereum JSON-RPC API: https://ethereum.org/en/developers/docs/apis/json-rpc/
