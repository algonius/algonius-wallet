---
title: 'Implement reject_transaction MCP Tool'
labels: ['enhancement', 'MCP', 'native-host', 'medium-priority', 'completed']
assignees: []
---

## Summary

Implement the `reject_transaction` MCP tool in the Native Host to enable AI Agents to reject pending transactions that should not be executed.

## Background

AI Agents need the ability to reject transactions that fail validation, are potentially malicious, or don't align with user preferences. This tool provides a secure way for AI to decline transaction execution while maintaining proper audit trails.

## Requirements

### Functional Requirements

- [x] Reject pending transactions by transaction ID
- [x] Support bulk rejection of multiple transactions
- [x] Provide rejection reasons for audit and logging
- [x] Maintain transaction rejection history
- [x] Send rejection notifications to relevant parties
- [x] Handle different rejection scenarios (security, policy, user preference)

### Technical Requirements

- [x] Create `native/pkg/mcp/tools/reject_transaction_tool.go`
- [x] Integrate with existing wallet manager and transaction tracking
- [x] Support multiple chain types (ETH, BSC initially)
- [x] Implement secure rejection mechanism with proper validation
- [x] Add comprehensive audit logging
- [x] Use structured response format for confirmation

### Security Requirements

- [x] Validate transaction ownership before allowing rejection
- [x] Authenticate AI agent requests to prevent unauthorized rejections
- [x] Implement rate limiting to prevent abuse
- [x] Ensure rejected transactions cannot be accidentally executed
- [x] Maintain secure audit trail of all rejection actions

## Acceptance Criteria

- [x] Tool can successfully reject pending transactions
- [x] Tool can reject multiple transactions in batch
- [x] Proper error handling for invalid transaction IDs or unauthorized access
- [x] Rejection reasons are properly logged and stored
- [x] Integration tests pass with various rejection scenarios
- [x] Security validations prevent unauthorized transaction rejections

## Implementation Details

### Files Created/Modified

- `native/pkg/mcp/tools/reject_transaction_tool.go` (new)
- `native/pkg/mcp/tools/reject_transaction_tool_test.go` (new)
- `native/pkg/wallet/manager.go` (extended with RejectTransactions method)
- `native/pkg/wallet/pending_transaction.go` (extended with rejection fields)
- `native/cmd/main.go` (registered new tool)
- `native/pkg/wallet/audit_log.go` (extended)

### API Schema

```json
{
  "name": "reject_transaction",
  "description": "Reject pending transactions by ID with specified reasons and optional notifications",
  "input_schema": {
    "type": "object",
    "properties": {
      "transaction_ids": { 
        "type": "string", 
        "description": "Comma-separated list of transaction hashes to reject" 
      },
      "reason": { 
        "type": "string", 
        "description": "Reason for rejection (e.g., 'suspicious_activity', 'high_gas_fee', 'user_request', 'security_concern', 'duplicate_transaction')" 
      },
      "details": { 
        "type": "string", 
        "description": "Additional details about the rejection reason" 
      },
      "notify_user": { 
        "type": "boolean", 
        "description": "Whether to send notification to user about the rejection (default: false)" 
      },
      "audit_log": { 
        "type": "boolean", 
        "description": "Whether to log the rejection in audit log (default: true)" 
      }
    },
    "required": ["transaction_ids", "reason"]
  },
  "output_schema": {
    "type": "object",
    "properties": {
      "summary": { 
        "type": "object",
        "properties": {
          "total_processed": { "type": "integer" },
          "successfully_rejected": { "type": "integer" },
          "failed_to_reject": { "type": "integer" }
        },
        "required": ["total_processed", "successfully_rejected", "failed_to_reject"]
      },
      "rejection_details": {
        "type": "object",
        "properties": {
          "reason": { "type": "string" },
          "details": { "type": "string" },
          "notify_user": { "type": "boolean" },
          "audit_log": { "type": "boolean" }
        }
      },
      "individual_results": {
        "type": "array",
        "items": {
          "type": "object",
          "properties": {
            "transaction_hash": { "type": "string" },
            "success": { "type": "boolean" },
            "error_message": { "type": "string" },
            "rejected_at": { "type": "string", "format": "date-time" },
            "audit_log_id": { "type": "string" }
          },
          "required": ["transaction_hash", "success"]
        }
      }
    },
    "required": ["summary", "rejection_details", "individual_results"]
  }
}
```

### Response Format

The tool returns a markdown-formatted response with:

1. Summary of transaction rejections (total processed, successfully rejected, failed to reject)
2. Rejection details (reason, details, notification and audit settings)
3. Individual results for each transaction (hash, success status, timestamp, audit log ID or error message)

### Key Features Implemented

1. **Bulk Transaction Rejection**: Accepts comma-separated list of transaction IDs
2. **Standardized Rejection Reasons**: Valid reasons include suspicious_activity, high_gas_fee, user_request, security_concern, duplicate_transaction
3. **Optional Notifications**: Can send user notifications when rejecting transactions
4. **Audit Logging**: Automatically logs all rejections with full details
5. **Comprehensive Error Handling**: Validates transaction IDs, ownership, and status before rejection
6. **Security Validation**: Ensures only authorized transactions are rejected
7. **Detailed Reporting**: Provides clear feedback on success and failure cases

## Dependencies

- Requires existing wallet manager and transaction tracking
- Depends on audit logging infrastructure
- Related to issue #002 (transaction confirmation)
- Related to issue #012 (get pending transactions)
- May depend on notification system for user alerts

## Testing Requirements

- [x] Unit tests for transaction rejection logic
- [x] Integration tests with real transaction scenarios
- [x] Test bulk rejection functionality
- [x] Test various rejection reasons and error cases
- [x] Security testing for unauthorized rejection attempts
- [x] Performance testing with large rejection batches

## Error Handling

The tool handles various error scenarios:

- Transaction not found or already processed
- Unauthorized access to transaction
- Invalid transaction ID format
- Network connectivity issues
- Database/storage errors
- Rate limiting exceeded

All errors are properly reported to the AI agent with descriptive messages.

## Audit and Logging

All rejection actions are logged with:

- Transaction ID and details
- Rejection reason and timestamp
- AI agent identity
- User notification status
- Chain and network information
- Unique audit log ID for traceability

The audit log is stored securely and can be used for compliance and security investigations.

## References

- Technical Spec: `docs/teck_spec.md`
- MCP API Documentation: `docs/apis/native_host_mcp_api.md`
- Related Implementation: `native/pkg/mcp/tools/confirm_transaction_tool.go`