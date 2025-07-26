# Issue 019: Implement Real-time Notifications

## Problem
No real-time notifications for wallet events. Users must manually check transaction status and balance changes.

## Current Behavior
- No push notifications for transaction confirmations
- No balance change alerts
- No network status updates
- Manual polling required for status updates

## Expected Behavior
- Real-time transaction confirmation notifications
- Balance change alerts
- Network connectivity status
- Configurable notification preferences

## Technical Details
- **New Files**: `native/pkg/events/notifier.go`, `native/pkg/sse/events.go`
- **Affected Files**: Transaction processing, wallet operations
- **Transport**: SSE (Server-Sent Events) for real-time updates
- **Storage**: Event subscription management

## Event Types
- **Transaction Events**: `transaction_confirmed`, `transaction_failed`, `transaction_pending`
- **Balance Events**: `balance_changed`, `token_received`, `token_sent`
- **Network Events**: `network_connected`, `network_disconnected`, `network_error`
- **Wallet Events**: `wallet_created`, `wallet_deleted`, `wallet_switched`

## SSE Event Format
```json
{
  "event": "transaction_confirmed",
  "data": {
    "tx_hash": "0x123...",
    "chain": "ethereum",
    "confirmations": 12,
    "block_number": 18500000,
    "timestamp": "2025-07-26T15:26:00Z"
  }
}
```

## New MCP Resource
- `notifications` - SSE endpoint for real-time notifications
- `notification_preferences` - Configure notification settings

## Acceptance Criteria
- [ ] SSE endpoint provides real-time notifications
- [ ] All transaction states trigger appropriate events
- [ ] Balance changes generate notifications
- [ ] Network status changes are communicated
- [ ] Configurable notification preferences
- [ ] Integration tests for all event types
- [ ] Browser extension receives notifications via Native Messaging

## Priority
Medium - Enhances user experience significantly

## Labels
feature, notifications, real-time, sse, medium-priority
