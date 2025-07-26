# Issue 021: Implement Transaction Simulation

## Problem
No transaction preview or simulation before execution. Users cannot see expected outcomes, gas costs, or potential failures before confirming transactions.

## Current Behavior
- Transactions execute immediately without preview
- No gas estimation provided
- No failure prediction
- No slippage warnings for swaps

## Expected Behavior
- Transaction simulation before execution
- Accurate gas estimation
- Failure prediction and warnings
- Slippage and price impact preview

## Technical Details
- **New Files**: `native/pkg/simulation/transaction.go`, `native/pkg/simulation/swap.go`
- **Affected Files**: `send_transaction`, `swap_tokens` tools
- **Simulation Engine**: Dry-run transactions using blockchain RPC
- **Gas Estimation**: Accurate gas limit and price calculations

## Simulation Features
- **Gas Estimation**: Predict exact gas usage and cost
- **Failure Detection**: Identify potential revert reasons
- **Price Impact**: Calculate slippage for swaps
- **Balance Check**: Verify sufficient funds
- **Nonce Management**: Check transaction ordering

## New MCP Tool
- `simulate_transaction` - Preview transaction outcomes

## Parameters
- Same as `send_transaction` but without execution
- Returns simulation results instead of transaction hash

## Simulation Response
```json
{
  "simulation": {
    "success": true,
    "gas_used": 21000,
    "gas_price": "20",
    "total_cost": "0.00042",
    "balance_change": "-0.50042",
    "warnings": ["High gas price detected"],
    "errors": []
  }
}
```

## Acceptance Criteria
- [ ] Accurate gas estimation for all transaction types
- [ ] Failure prediction with specific error messages
- [ ] Slippage calculation for token swaps
- [ ] Balance verification before execution
- [ ] Integration tests for simulation accuracy
- [ ] Browser extension displays simulation results

## Priority
High - Critical for user safety and experience

## Labels
feature, simulation, safety, gas-estimation, high-priority
