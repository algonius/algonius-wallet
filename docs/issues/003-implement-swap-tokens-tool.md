---
title: 'Implement swap_tokens MCP Tool'
labels: ['enhancement', 'MCP', 'native-host', 'defi', 'medium-priority']
assignees: []
---

## Summary

Implement the `swap_tokens` MCP tool to enable AI Agents to perform token swaps through decentralized exchanges (DEX).

## Background

Token swapping is a core DeFi operation for AI trading strategies. This tool will integrate with popular DEX protocols to enable automated token exchanges with slippage protection and price validation.

## Requirements

### Functional Requirements

- [ ] Support Uniswap V2/V3 protocol integration
- [ ] Implement token-to-token swaps on Ethereum and BSC
- [ ] Add slippage protection and price impact calculation
- [ ] Support exact input and exact output swap modes
- [ ] Implement deadline protection for transactions
- [ ] Add multi-hop routing for optimal prices

### Technical Requirements

- [ ] Create `native/pkg/mcp/tools/swap_tokens_tool.go`
- [ ] Integrate with Uniswap V2/V3 smart contracts
- [ ] Implement DEX router interface abstraction
- [ ] Add price quote functionality before swap execution
- [ ] Support multiple DEX protocols (Uniswap, PancakeSwap)
- [ ] Implement gas estimation for swap transactions

### Security Requirements

- [ ] Validate token contract addresses
- [ ] Implement maximum slippage limits
- [ ] Add price impact warnings for large trades
- [ ] Validate token amounts and decimals
- [ ] Implement transaction deadline protection
- [ ] Add approval transaction handling for ERC-20 tokens

## Acceptance Criteria

- [ ] Tool can successfully swap tokens on Uniswap
- [ ] Slippage protection works correctly
- [ ] Price quotes are accurate within acceptable margins
- [ ] Gas estimation is reliable
- [ ] Integration tests pass with real DEX interactions
- [ ] Security validations prevent malicious swaps

## Implementation Details

### Files to Create/Modify

- `native/pkg/mcp/tools/swap_tokens_tool.go` (new)
- `native/pkg/wallet/dex/` (new package)
- `native/pkg/wallet/dex/uniswap.go` (new)
- `native/pkg/wallet/dex/interfaces.go` (new)
- `native/cmd/main.go` (register tool)

### API Schema

```json
{
  "name": "swap_tokens",
  "description": "Swap tokens through DEX protocols",
  "inputSchema": {
    "type": "object",
    "properties": {
      "chain": { "type": "string", "enum": ["ethereum", "bsc"] },
      "token_in": { "type": "string", "description": "Input token contract address" },
      "token_out": { "type": "string", "description": "Output token contract address" },
      "amount_in": { "type": "string", "description": "Input amount (for exact input)" },
      "amount_out": { "type": "string", "description": "Output amount (for exact output)" },
      "slippage_tolerance": {
        "type": "number",
        "default": 0.5,
        "description": "Max slippage in percent"
      },
      "deadline": {
        "type": "number",
        "default": 300,
        "description": "Transaction deadline in seconds"
      },
      "dex": { "type": "string", "enum": ["uniswap", "pancakeswap"], "default": "uniswap" }
    },
    "required": ["chain", "token_in", "token_out"],
    "oneOf": [{ "required": ["amount_in"] }, { "required": ["amount_out"] }]
  }
}
```

### Response Format

```json
{
  "transaction_hash": "0x...",
  "amount_in": "1000000000000000000",
  "amount_out": "2500000000",
  "price_impact": 0.15,
  "gas_used": "150000",
  "dex_used": "uniswap_v3",
  "route": ["0xA0b86a33E6441e...", "0xC02aaA39b223..."]
}
```

## Dependencies

- Requires DEX smart contract ABIs and addresses
- Depends on go-ethereum library for contract interaction
- Related to issue #001 (send_transaction tool)
- Requires price oracle integration for validation

## Testing Requirements

- [ ] Unit tests for swap calculation logic
- [ ] Integration tests with Uniswap testnet
- [ ] Mock tests for contract interaction
- [ ] Slippage protection validation tests
- [ ] Gas estimation accuracy tests

## DEX Integration Priority

1. **Phase 1**: Uniswap V2 on Ethereum
2. **Phase 2**: PancakeSwap on BSC
3. **Phase 3**: Uniswap V3 with concentrated liquidity
4. **Phase 4**: Additional DEX protocols

## Configuration

- [ ] DEX contract addresses per chain
- [ ] Default slippage tolerances
- [ ] Maximum trade size limits
- [ ] Supported token whitelist

## References

- Uniswap V2 Documentation: https://docs.uniswap.org/contracts/v2/overview
- Uniswap V3 Documentation: https://docs.uniswap.org/contracts/v3/overview
- Technical Spec: `docs/teck_spec.md`
- MCP API Documentation: `docs/apis/native_host_mcp_api.md`
