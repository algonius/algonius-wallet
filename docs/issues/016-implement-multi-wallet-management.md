# Issue 016: Implement Multi-Wallet Management

## Problem
Current wallet system only supports a single active wallet with no management capabilities. Users cannot list, switch between, or manage multiple wallets.

## Current Behavior
- Only one wallet per chain can be created
- No way to view all created wallets
- No wallet switching functionality
- No wallet deletion or labeling

## Expected Behavior
- Create multiple wallets per chain
- List all wallets with metadata (address, label, creation date)
- Switch active wallet for operations
- Label/nickname wallets for easy identification
- Delete wallets (with confirmation)

## Technical Details
- **New Files**: `native/pkg/wallet/manager.go`, `native/pkg/mcp/wallet_management.go`
- **Affected Files**: `native/pkg/mcp/tools.go`, wallet storage layer
- **Storage**: Extend wallet storage to support multiple wallets per chain
- **APIs**: New MCP tools for wallet management

## New MCP Tools Required
- `list_wallets` - List all wallets across chains
- `switch_wallet` - Change active wallet
- `label_wallet` - Add/edit wallet labels
- `delete_wallet` - Remove wallet (with confirmation)

## Acceptance Criteria
- [ ] Can create multiple wallets per chain
- [ ] `list_wallets` returns all wallets with metadata
- [ ] `switch_wallet` changes active wallet for operations
- [ ] Wallet labels persist across sessions
- [ ] Delete wallet requires confirmation and cleanup
- [ ] Integration tests for all wallet management operations

## Priority
High - Core user experience feature

## Labels
feature, wallet-management, user-experience, high-priority
