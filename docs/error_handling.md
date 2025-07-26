# Error Handling in Algonius Wallet

This document describes the standardized error handling mechanism implemented in the Algonius Wallet.

## Overview

The Algonius Wallet implements a consistent error handling approach across all MCP tools. All errors follow a standardized format that includes:

1. **Error Code**: A unique identifier for the error type
2. **Message**: A clear, human-readable description of the error
3. **Details**: Additional context about the error (optional)
4. **Suggestion**: Actionable guidance for resolving the issue (optional)

## Error Categories

The error handling system supports several error categories:

1. **Validation Errors**: Invalid parameters or missing required fields
2. **Network Errors**: Connection issues, timeouts, or RPC failures
3. **Wallet Errors**: Issues related to wallet operations (balance, addresses, etc.)
4. **Token Errors**: Problems with token contracts or unsupported tokens
5. **Permission Errors**: Unauthorized operations
6. **General Errors**: Internal errors or unexpected conditions

## Error Response Format

All errors are returned in a consistent format:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Clear description of the error",
    "details": "Additional context (optional)",
    "suggestion": "Actionable guidance (optional)"
  }
}
```

## Implementation

The error handling is implemented in the `pkg/errors` package, which provides:

1. Standard error codes for common error types
2. Helper functions for creating specific types of errors
3. A standardized `Error` struct that implements the `error` interface
4. Methods for adding details and suggestions to errors

All MCP tool handlers have been updated to use this standardized error handling approach.

## Examples

Here are some examples of how errors are handled:

### Validation Error
```
{
  "error": {
    "code": "INVALID_PARAMETER",
    "message": "Invalid 'address' parameter",
    "details": "Address format is invalid for Ethereum chain",
    "suggestion": "Provide a valid Ethereum address"
  }
}
```

### Network Error
```
{
  "error": {
    "code": "NETWORK_CONNECTION_ERROR",
    "message": "Failed to connect during balance query: connection refused",
    "details": "connection refused",
    "suggestion": "Check your internet connection and try again"
  }
}
```

### Insufficient Balance Error
```
{
  "error": {
    "code": "INSUFFICIENT_BALANCE",
    "message": "Insufficient balance for token 'ETH'",
    "details": "Current balance: 0.1 ETH, Required: 0.5 ETH",
    "suggestion": "Add more funds to your wallet or use a different token"
  }
}
```

## Benefits

This standardized error handling approach provides several benefits:

1. **Consistency**: All errors follow the same format regardless of which tool generated them
2. **Clarity**: Error messages are clear and specific
3. **Actionability**: Suggestions help users resolve issues
4. **Debugging**: Error codes and details help developers diagnose problems
5. **User Experience**: Better error messages lead to a better user experience