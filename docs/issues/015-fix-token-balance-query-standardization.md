# Issue 015: Fix Token Balance Query Standardization

## Problem
Token balance queries for BSC and Solana native tokens fail with "unsupported token" errors. The current implementation only supports ETH for Ethereum, but lacks standardized token identifiers across chains.

## Current Behavior
- `get_balance` with "BNB" or "SOL" returns "unsupported token"
- BSC native token queries require specific contract addresses
- No consistent cross-chain token identification

## Expected Behavior
- Support native tokens: ETH (Ethereum), BNB (BSC), SOL (Solana)
- Standardized token identifiers across all supported chains
- Clear error messages for unsupported tokens

## Technical Details
- **Affected Files**: `native/pkg/mcp/tools.go`, `native/pkg/wallet/chain.go`
- **Related Components**: Balance query handlers, chain-specific implementations
- **Testing**: Add integration tests for BSC and Solana balance queries

## Acceptance Criteria
- [ ] `get_balance` accepts "ETH", "BNB", "SOL" as valid token identifiers
- [ ] All supported chains return correct native token balances
- [ ] Integration tests pass for all chain/token combinations
- [ ] Backward compatibility maintained for contract addresses

## Priority
High - Critical for multi-chain functionality

## Labels
bug, multi-chain, balance-query, high-priority
