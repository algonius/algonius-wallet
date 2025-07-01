---
title: 'Implement switch_chain MCP Tool'
labels: ['enhancement', 'MCP', 'native-host', 'multi-chain', 'low-priority']
assignees: []
---

## Summary

Implement the `switch_chain` MCP tool to enable AI Agents to dynamically switch between supported blockchain networks during runtime.

## Background

AI trading agents need the ability to switch between different blockchain networks (Ethereum, BSC, Solana) to access different DeFi protocols and opportunities without restarting the native host.

## Requirements

### Functional Requirements

- [ ] Switch active blockchain network at runtime
- [ ] Validate chain availability before switching
- [ ] Update all dependent services (RPC, explorer APIs)
- [ ] Maintain wallet compatibility across chains
- [ ] Provide chain status and connection health
- [ ] Support graceful fallback for failed switches

### Technical Requirements

- [ ] Create `native/pkg/mcp/tools/switch_chain_tool.go`
- [ ] Extend chain factory for dynamic switching
- [ ] Update wallet manager for multi-chain state
- [ ] Implement connection pooling per chain
- [ ] Add chain validation and health checks
- [ ] Update existing tools to respect active chain

### State Management

- [ ] Track active chain per session
- [ ] Persist chain preference across restarts
- [ ] Handle concurrent chain operations
- [ ] Validate operations against active chain
- [ ] Update resource URIs for chain context

## Acceptance Criteria

- [ ] Tool can successfully switch between supported chains
- [ ] All dependent services update correctly
- [ ] Wallet operations work on new chain
- [ ] Chain validation prevents invalid switches
- [ ] Integration tests pass for all chain combinations
- [ ] Performance impact is minimal

## Implementation Details

### Files to Create/Modify

- `native/pkg/mcp/tools/switch_chain_tool.go` (new)
- `native/pkg/wallet/chain/factory.go` (extend)
- `native/pkg/wallet/manager.go` (extend)
- `native/pkg/config/chains.go` (new)
- `native/cmd/main.go` (register tool)

### API Schema

```json
{
  "name": "switch_chain",
  "description": "Switch to a different blockchain network",
  "inputSchema": {
    "type": "object",
    "properties": {
      "chain": {
        "type": "string",
        "enum": ["ethereum", "bsc", "solana"],
        "description": "Target blockchain network"
      },
      "rpc_url": { "type": "string", "description": "Custom RPC URL (optional)" },
      "validate_connection": {
        "type": "boolean",
        "default": true,
        "description": "Test connection before switching"
      }
    },
    "required": ["chain"]
  }
}
```

### Response Format

```json
{
  "previous_chain": "ethereum",
  "current_chain": "bsc",
  "chain_id": 56,
  "rpc_url": "https://bsc-dataseed.binance.org/",
  "block_height": 34567890,
  "connection_status": "healthy",
  "switch_time": "2025-06-24T07:00:00Z"
}
```

### Chain Configuration

```json
{
  "chains": {
    "ethereum": {
      "name": "Ethereum Mainnet",
      "chain_id": 1,
      "rpc_urls": ["https://mainnet.infura.io/v3/...", "https://eth-mainnet.alchemyapi.io/v2/..."],
      "explorer_api": "https://api.etherscan.io/api",
      "native_token": "ETH",
      "supports": ["evm", "erc20", "uniswap"]
    },
    "bsc": {
      "name": "Binance Smart Chain",
      "chain_id": 56,
      "rpc_urls": ["https://bsc-dataseed.binance.org/"],
      "explorer_api": "https://api.bscscan.com/api",
      "native_token": "BNB",
      "supports": ["evm", "bep20", "pancakeswap"]
    }
  }
}
```

## Dependencies

- Requires updated chain factory architecture
- Related to all MCP tools that interact with blockchain
- May need configuration file updates

## Testing Requirements

- [ ] Unit tests for chain switching logic
- [ ] Integration tests for each supported chain
- [ ] Connection failure and recovery tests
- [ ] Concurrent operation tests
- [ ] State persistence tests

## Impact on Existing Tools

All existing MCP tools need updates to respect active chain:

- [ ] `get_balance` - Query balance on active chain
- [ ] `send_transaction` - Send on active chain
- [ ] `swap_tokens` - Use DEX on active chain
- [ ] `get_transactions` - Query history on active chain
- [ ] Resource providers - Update chain context

## Chain Health Monitoring

- [ ] Regular connection health checks
- [ ] Automatic failover to backup RPC URLs
- [ ] Block height synchronization monitoring
- [ ] Network latency tracking
- [ ] Connection pool optimization

## Configuration Management

- [ ] Default chain preference
- [ ] RPC URL priority lists
- [ ] Health check intervals
- [ ] Connection timeouts
- [ ] Fallback strategies

## Security Considerations

- [ ] Validate chain IDs to prevent attacks
- [ ] Verify RPC URL authenticity
- [ ] Rate limit chain switching operations
- [ ] Audit log for chain switches
- [ ] Prevent concurrent unsafe operations

## Performance Optimizations

- [ ] Connection pooling per chain
- [ ] Lazy initialization of inactive chains
- [ ] Cached chain metadata
- [ ] Efficient state synchronization
- [ ] Minimal switching overhead

## Future Enhancements

- [ ] Support for custom/testnet chains
- [ ] Multi-chain operations in parallel
- [ ] Cross-chain bridge integration
- [ ] Advanced routing algorithms
- [ ] Chain analytics and recommendations

## References

- Chain Registry: https://github.com/cosmos/chain-registry
- EIP-155: https://eips.ethereum.org/EIPS/eip-155
- Technical Spec: `docs/teck_spec.md`
- MCP API Documentation: `docs/apis/native_host_mcp_api.md`
