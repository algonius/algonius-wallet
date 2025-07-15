---
title: 'Implement reject_transaction MCP Tool'
labels: ['enhancement', 'MCP', 'native-host', 'medium-priority']
assignees: []
---

## Summary

Implement the `reject_transaction` MCP tool in the Native Host to enable AI Agents to reject pending transactions that should not be executed.

## Background

AI Agents need the ability to reject transactions that fail validation, are potentially malicious, or don't align with user preferences. This tool provides a secure way for AI to decline transaction execution while maintaining proper audit trails.

## Requirements

### Functional Requirements

- [ ] Reject pending transactions by transaction ID
- [ ] Support bulk rejection of multiple transactions
- [ ] Provide rejection reasons for audit and logging
- [ ] Maintain transaction rejection history
- [ ] Send rejection notifications to relevant parties
- [ ] Handle different rejection scenarios (security, policy, user preference)

### Technical Requirements

- [ ] Create `native/pkg/mcp/tools/reject_transaction_tool.go`
- [ ] Integrate with existing wallet manager and transaction tracking
- [ ] Support multiple chain types (ETH, BSC initially)
- [ ] Implement secure rejection mechanism with proper validation
- [ ] Add comprehensive audit logging
- [ ] Use structured response format for confirmation

### Security Requirements

- [ ] Validate transaction ownership before allowing rejection
- [ ] Authenticate AI agent requests to prevent unauthorized rejections
- [ ] Implement rate limiting to prevent abuse
- [ ] Ensure rejected transactions cannot be accidentally executed
- [ ] Maintain secure audit trail of all rejection actions

## Acceptance Criteria

- [ ] Tool can successfully reject pending transactions
- [ ] Tool can reject multiple transactions in batch
- [ ] Proper error handling for invalid transaction IDs or unauthorized access
- [ ] Rejection reasons are properly logged and stored
- [ ] Integration tests pass with various rejection scenarios
- [ ] Security validations prevent unauthorized transaction rejections

## Implementation Details

### Files to Create/Modify

- `native/pkg/mcp/tools/reject_transaction_tool.go` (new)
- `native/pkg/wallet/transaction.go` (extend)
- `native/pkg/wallet/chain/eth_chain.go` (extend)
- `native/pkg/mcp/server.go` (register new tool)
- `native/pkg/wallet/audit_log.go` (extend)

### API Schema

```json
{
  "name": "reject_transaction",
  "description": "Reject one or more pending transactions",
  "inputSchema": {
    "type": "object",
    "properties": {
      "transaction_ids": {
        "type": "array",
        "items": { "type": "string" },
        "description": "Array of transaction IDs to reject"
      },
      "reason": {
        "type": "string",
        "enum": ["security_risk", "policy_violation", "user_preference", "invalid_parameters", "insufficient_funds", "other"],
        "description": "Reason for rejection"
      },
      "details": {
        "type": "string",
        "description": "Additional details about the rejection (optional)"
      },
      "notify_user": {
        "type": "boolean",
        "default": true,
        "description": "Whether to notify user about rejection"
      }
    },
    "required": ["transaction_ids", "reason"]
  }
}
```

### Response Format

```json
{
  "rejected_transactions": [
    {
      "transaction_id": "0x123...",
      "status": "rejected",
      "reason": "security_risk",
      "details": "Suspicious recipient address detected",
      "rejected_at": "2024-01-01T12:00:00Z",
      "chain": "ethereum"
    }
  ],
  "failed_rejections": [
    {
      "transaction_id": "0x456...",
      "error": "Transaction not found or already processed",
      "error_code": "NOT_FOUND"
    }
  ],
  "total_processed": 2,
  "total_rejected": 1,
  "total_failed": 1
}
```

## Dependencies

- Requires existing wallet manager and transaction tracking
- Depends on audit logging infrastructure
- Related to issue #002 (transaction confirmation)
- Related to issue #012 (get pending transactions)
- May depend on notification system for user alerts

## Testing Requirements

- [ ] Unit tests for transaction rejection logic
- [ ] Integration tests with real transaction scenarios
- [ ] Test bulk rejection functionality
- [ ] Test various rejection reasons and error cases
- [ ] Security testing for unauthorized rejection attempts
- [ ] Performance testing with large rejection batches

## Error Handling

The tool should handle various error scenarios:

- Transaction not found or already processed
- Unauthorized access to transaction
- Invalid transaction ID format
- Network connectivity issues
- Database/storage errors
- Rate limiting exceeded

## Audit and Logging

All rejection actions should be logged with:
- Transaction ID and details
- Rejection reason and timestamp
- AI agent identity
- User notification status
- Chain and network information

## References

- Technical Spec: `docs/teck_spec.md`
- MCP API Documentation: `docs/apis/native_host_mcp_api.md`
- Related Implementation: `native/pkg/mcp/tools/confirm_transaction_tool.go`