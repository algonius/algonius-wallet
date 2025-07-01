---
title: 'Implement get_transactions MCP Tool'
labels: ['enhancement', 'MCP', 'native-host', 'medium-priority']
assignees: []
---

## Summary

Implement the `get_transactions` MCP tool to allow AI Agents to query transaction history for wallet addresses across supported blockchains.

## Background

Transaction history is essential for AI trading analysis and portfolio tracking. This tool provides comprehensive transaction data with filtering and pagination capabilities.

## Requirements

### Functional Requirements

- [ ] Retrieve transaction history for wallet addresses
- [ ] Support filtering by date range, transaction type, and token
- [ ] Implement pagination for large transaction sets
- [ ] Parse and categorize transaction types (send, receive, swap, etc.)
- [ ] Include transaction details (amount, gas, status, timestamp)
- [ ] Support multiple blockchain networks (ETH, BSC, Solana)

### Technical Requirements

- [ ] Create `native/pkg/mcp/tools/get_transactions_tool.go`
- [ ] Integrate with blockchain explorers APIs (Etherscan, BSCScan)
- [ ] Implement transaction parsing and categorization
- [ ] Add caching for frequently requested data
- [ ] Support concurrent requests for multiple addresses

### Performance Requirements

- [ ] Response time under 3 seconds for cached queries
- [ ] Support pagination with configurable page sizes
- [ ] Implement rate limiting for API calls
- [ ] Cache transaction data with appropriate TTL

## Acceptance Criteria

- [ ] Tool returns accurate transaction history
- [ ] Filtering and pagination work correctly
- [ ] Transaction categorization is accurate
- [ ] Supports all target blockchain networks
- [ ] Integration tests pass with real wallet data
- [ ] Performance requirements met

## Implementation Details

### Files to Create/Modify

- `native/pkg/mcp/tools/get_transactions_tool.go` (new)
- `native/pkg/wallet/explorer/` (new package)
- `native/pkg/wallet/explorer/etherscan.go` (new)
- `native/pkg/wallet/explorer/interfaces.go` (new)
- `native/cmd/main.go` (register tool)

### API Schema

```json
{
  "name": "get_transactions",
  "description": "Get transaction history for a wallet address",
  "inputSchema": {
    "type": "object",
    "properties": {
      "chain": { "type": "string", "enum": ["ethereum", "bsc", "solana"] },
      "address": { "type": "string", "description": "Wallet address" },
      "limit": {
        "type": "number",
        "default": 50,
        "maximum": 100,
        "description": "Number of transactions to return"
      },
      "offset": { "type": "number", "default": 0, "description": "Pagination offset" },
      "from_date": { "type": "string", "format": "date-time", "description": "Start date filter" },
      "to_date": { "type": "string", "format": "date-time", "description": "End date filter" },
      "tx_type": {
        "type": "string",
        "enum": ["all", "send", "receive", "swap", "approve"],
        "default": "all"
      },
      "token": { "type": "string", "description": "Filter by specific token contract address" }
    },
    "required": ["chain", "address"]
  }
}
```

### Response Format

```json
{
  "transactions": [
    {
      "hash": "0x...",
      "type": "send",
      "from": "0x...",
      "to": "0x...",
      "amount": "1000000000000000000",
      "token": {
        "address": "0x...",
        "symbol": "USDT",
        "decimals": 6
      },
      "gas_used": "21000",
      "gas_price": "20000000000",
      "block_number": 18500000,
      "timestamp": "2025-06-24T07:00:00Z",
      "status": "success"
    }
  ],
  "total_count": 1250,
  "has_more": true
}
```

## Dependencies

- Requires blockchain explorer APIs (Etherscan, BSCScan)
- May need API keys for rate limit increases
- Related to wallet management and chain interfaces

## Testing Requirements

- [ ] Unit tests for transaction parsing
- [ ] Integration tests with real blockchain data
- [ ] Mock tests for API failures
- [ ] Pagination functionality tests
- [ ] Filter accuracy tests

## API Integration

- [ ] Etherscan API for Ethereum transactions
- [ ] BSCScan API for BSC transactions
- [ ] Solana RPC API for Solana transactions
- [ ] Rate limiting and error handling
- [ ] API key management

## Transaction Categories

- **Send**: Outgoing transfers from wallet
- **Receive**: Incoming transfers to wallet
- **Swap**: DEX token exchanges
- **Approve**: Token approval transactions
- **Contract**: Smart contract interactions
- **NFT**: NFT transfers and mints

## Configuration

- [ ] Explorer API endpoints per chain
- [ ] API keys and rate limits
- [ ] Cache TTL settings
- [ ] Default pagination sizes

## References

- Etherscan API: https://docs.etherscan.io/
- BSCScan API: https://docs.bscscan.com/
- Technical Spec: `docs/teck_spec.md`
- MCP API Documentation: `docs/apis/native_host_mcp_api.md`
