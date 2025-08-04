### 1.8 reject_transaction

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
  },
  "error_schema": {
    "type": "object",
    "properties": {
      "code": { "type": "integer" },
      "message": { "type": "string" }
    },
    "required": ["code", "message"]
  },
  "security": "需用户授权"
}
```