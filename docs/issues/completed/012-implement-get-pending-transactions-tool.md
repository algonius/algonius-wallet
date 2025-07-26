---
title: 'Implement get_pending_transactions MCP Tool'
labels: ['enhancement', 'MCP', 'native-host', 'medium-priority', 'completed']
assignees: []
---

## Summary

Implement the `get_pending_transactions` MCP tool in the Native Host to enable AI Agents to query and monitor pending transactions that require confirmation.

## Background

AI Agents need the ability to query pending transactions to make intelligent decisions about which transactions to confirm or reject. This tool will provide visibility into transactions that are waiting for confirmation, along with their details and current status.

## Requirements

### Functional Requirements

- [x] Query all pending transactions across supported chains
- [x] Filter pending transactions by chain, address, or transaction type
- [x] Provide detailed transaction information including amount, recipient, gas fees
- [x] Support pagination for large result sets
- [x] Include transaction priority/urgency indicators
- [x] Provide estimated confirmation time and gas price recommendations

### Technical Requirements

- [x] Create `native/pkg/mcp/tools/get_pending_transactions_tool.go`
- [x] Integrate with existing wallet manager and chain interfaces
- [x] Support multiple chain types (ETH, BSC initially)
- [x] Implement efficient transaction status tracking
- [x] Add comprehensive filtering and sorting options
- [x] Use structured response format for easy AI processing

### Security Requirements

- [x] Only return transactions for addresses controlled by current wallet
- [x] Validate request parameters to prevent information leakage
- [x] Implement rate limiting to prevent abuse
- [x] Ensure sensitive transaction details are properly handled

## Acceptance Criteria

- [x] Tool can successfully query pending transactions on Ethereum
- [x] Tool can filter results by chain, address, and transaction type
- [x] Proper error handling for network issues and invalid requests
- [x] Response includes all necessary transaction details for AI decision making
- [x] Integration tests pass with various filter combinations
- [x] Performance is acceptable with large numbers of pending transactions

## Implementation Details

### Files Created/Modified

- `native/pkg/mcp/tools/get_pending_transactions_tool.go` (new)
- `native/pkg/mcp/tools/get_pending_transactions_tool_test.go` (new)
- `native/pkg/wallet/pending_transaction.go` (new)
- `native/pkg/wallet/manager.go` (extended with GetPendingTransactions method)
- `native/cmd/main.go` (registered new tool)

### API Schema

```json
{
  "name": "get_pending_transactions",
  "description": "Query pending transactions that require confirmation",
  "input_schema": {
    "type": "object",
    "properties": {
      "chain": { 
        "type": "string", 
        "description": "Filter by specific chain (ethereum, bsc, solana) (optional)"
      },
      "address": { 
        "type": "string", 
        "description": "Filter by specific address (from or to) (optional)" 
      },
      "transaction_type": { 
        "type": "string", 
        "enum": ["transfer", "swap", "contract"],
        "description": "Filter by transaction type (optional)"
      },
      "limit": { 
        "type": "integer", 
        "default": 10,
        "maximum": 100,
        "description": "Maximum number of results to return (1-100)" 
      },
      "offset": { 
        "type": "integer", 
        "default": 0,
        "minimum": 0,
        "description": "Number of results to skip for pagination" 
      }
    },
    "required": []
  },
  "output_schema": {
    "type": "object",
    "properties": {
      "transactions": {
        "type": "array",
        "items": {
          "type": "object",
          "properties": {
            "hash": { "type": "string" },
            "chain": { "type": "string" },
            "from": { "type": "string" },
            "to": { "type": "string" },
            "amount": { "type": "string" },
            "token": { "type": "string" },
            "type": { "type": "string", "enum": ["transfer", "swap", "contract"] },
            "status": { "type": "string", "enum": ["pending", "confirmed", "failed"] },
            "confirmations": { "type": "integer" },
            "required_confirmations": { "type": "integer" },
            "block_number": { "type": "integer" },
            "nonce": { "type": "integer" },
            "gas_fee": { "type": "string" },
            "priority": { "type": "string", "enum": ["low", "medium", "high"] },
            "estimated_confirmation_time": { "type": "string" },
            "submitted_at": { "type": "string", "format": "date-time" },
            "last_checked": { "type": "string", "format": "date-time" }
          },
          "required": ["hash", "chain", "from", "to", "amount", "token", "type", "status"]
        }
      },
      "total_count": { "type": "integer" },
      "has_more": { "type": "boolean" }
    },
    "required": ["transactions", "total_count", "has_more"]
  }
}
```

### Response Format

The tool returns a JSON object with:

1. Array of pending transactions with full details
2. Total count of matching transactions
3. Boolean indicating if there are more results available

### Key Features Implemented

1. **Multi-chain Support**: Works with Ethereum, BSC, and Solana
2. **Advanced Filtering**: Filter by chain, address (from/to), and transaction type
3. **Pagination**: Support for large result sets with limit/offset
4. **Rich Transaction Details**: All relevant information for AI decision making
5. **Comprehensive Error Handling**: Validates parameters and handles edge cases
6. **Security Validation**: Only returns transactions for current wallet

## Dependencies

- Requires existing wallet manager and chain interfaces
- Depends on transaction tracking infrastructure
- Related to issue #002 (transaction confirmation)
- Related to issue #013 (transaction rejection)

## Testing Requirements

- [x] Unit tests for transaction querying and filtering
- [x] Integration tests with real blockchain interaction
- [x] Test pagination and large result sets
- [x] Test various filter combinations
- [x] Performance testing with high transaction volumes

## References

- Technical Spec: `docs/teck_spec.md`
- MCP API Documentation: `docs/apis/native_host_mcp_api.md`
- Related Implementation: `native/pkg/mcp/tools/get_transactions_tool.go`