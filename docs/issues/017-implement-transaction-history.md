# Issue 017: Implement Transaction History

## Problem
No way to query historical transactions for a wallet. Current system only shows pending transactions, lacking complete transaction history.

## Current Behavior
- Only pending transactions visible via `get_pending_transactions`
- No historical transaction data available
- No filtering or pagination for transaction lists

## Expected Behavior
- Complete transaction history per wallet
- Filter by date range, transaction type, status
- Pagination support for large histories
- Export functionality for transaction records

## Technical Details
- **New Files**: `native/pkg/mcp/transaction_history.go`
- **Affected Files**: `native/pkg/mcp/tools.go`, transaction storage layer
- **Storage**: Extend transaction storage to persist historical data
- **APIs**: New MCP tool for transaction history

## New MCP Tool Required
- `get_transaction_history` - Query historical transactions with filtering

## Parameters
- `address` - Wallet address (optional, defaults to active wallet)
- `chain` - Blockchain to query (optional, all chains)
- `from_date` - Start date for filtering
- `to_date` - End date for filtering
- `type` - Transaction type (transfer, swap, contract, etc.)
- `status` - Transaction status (confirmed, failed, pending)
- `limit` - Number of results (default: 50, max: 1000)
- `offset` - Pagination offset

## Acceptance Criteria
- [ ] Historical transactions persist across sessions
- [ ] `get_transaction_history` returns complete transaction records
- [ ] Filtering works for all specified parameters
- [ ] Pagination handles large datasets efficiently
- [ ] Export functionality (CSV/JSON format)
- [ ] Integration tests for all filtering and pagination scenarios

## Priority
Medium - Important for user transparency and accounting

## Labels
feature, transaction-history, transparency, medium-priority
