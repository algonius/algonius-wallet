# Issue 018: Enhance Error Handling and User Messages

## Problem
Current error messages are generic and unhelpful. Users receive vague errors like "failed to get balance" without understanding what went wrong or how to fix it.

## Current Behavior
- Generic error messages: "failed to get balance", "unsupported token"
- No error codes or categorization
- No actionable guidance for users
- Inconsistent error formats across tools

## Expected Behavior
- Specific error codes and categories
- Clear, actionable error messages
- Suggested solutions for common issues
- Consistent error format across all tools

## Technical Details
- **Affected Files**: All MCP tool handlers in `native/pkg/mcp/`
- **New Files**: `native/pkg/errors/errors.go`, error message localization
- **Components**: Error handling middleware, validation layer
- **Format**: Standardized error response structure

## Error Categories
- **Validation Errors**: Invalid parameters, missing required fields
- **Network Errors**: Connection issues, timeout, RPC failures
- **Wallet Errors**: Insufficient balance, invalid address
- **Token Errors**: Unsupported token, invalid contract address
- **Permission Errors**: Unauthorized operations

## Error Response Format
```json
{
  "error": {
    "code": "TOKEN_NOT_SUPPORTED",
    "message": "Token 'BNB' is not supported on Ethereum chain",
    "details": "Use 'bsc' chain parameter for BNB token",
    "suggestion": "Try: get_balance with chain='bsc' and token='BNB'"
  }
}
```

## Acceptance Criteria
- [ ] All error messages include specific error codes
- [ ] Error messages provide actionable guidance
- [ ] Consistent error format across all MCP tools
- [ ] Error messages are user-friendly and helpful
- [ ] Integration tests verify error scenarios
- [ ] Documentation includes error code reference

## Priority
High - Critical for user experience and debugging

## Labels
enhancement, error-handling, user-experience, high-priority
