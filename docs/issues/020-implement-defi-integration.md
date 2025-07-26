# Issue 020: Implement DeFi Integration

## Problem
Limited DeFi functionality with only basic token swaps. No access to DeFi protocols, yield farming, or advanced trading features.

## Current Behavior
- Basic token swaps via single DEX
- No price comparison across DEXes
- No liquidity pool information
- No yield farming opportunities

## Expected Behavior
- DEX aggregator for best rates
- Liquidity pool information and analytics
- Yield farming opportunities tracking
- Advanced DeFi protocol integration

## Technical Details
- **New Files**: `native/pkg/defi/aggregator.go`, `native/pkg/defi/protocols/`
- **Affected Files**: `native/pkg/mcp/swap_tokens.go`, swap functionality
- **Protocols**: Uniswap V3, PancakeSwap, SushiSwap, Curve, Aave
- **APIs**: Price feeds, liquidity data, yield rates

## Enhanced Swap Tool
Extend `swap_tokens` with:
- **DEX Selection**: `auto`, `uniswap_v3`, `pancakeswap`, `sushiswap`
- **Price Comparison**: Best rate across multiple DEXes
- **Slippage Protection**: Dynamic slippage based on market conditions
- **Route Optimization**: Multi-hop routing for better rates

## New MCP Tools
- `get_liquidity_pools` - Query available liquidity pools
- `get_yield_opportunities` - List yield farming options
- `get_defi_protocols` - Available DeFi protocols per chain
- `get_token_price` - Real-time token prices across DEXes

## Parameters for Enhanced Swap
```json
{
  "dex": "auto",
  "price_comparison": true,
  "max_slippage": "auto",
  "route_optimization": true
}
```

## Acceptance Criteria
- [ ] DEX aggregator finds best rates across protocols
- [ ] Liquidity pool information accessible via MCP
- [ ] Yield farming opportunities tracked and updated
- [ ] Enhanced swap with route optimization
- [ ] Real-time price feeds from multiple sources
- [ ] Integration tests for all DeFi operations
- [ ] Security validation for DeFi interactions

## Priority
Medium - Advanced feature for power users

## Labels
feature, defi, dex, yield-farming, medium-priority
