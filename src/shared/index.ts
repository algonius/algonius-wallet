/**
 * Shared utilities and types for the extension
 */

export enum McpErrorCode {
  HOST_NOT_FOUND = 1001,
  CONNECTION_FAILED = 1002,
  REQUEST_TIMEOUT = 1003,
  INVALID_RESPONSE = 1004,
  UNKNOWN_ERROR = 1999,
}

export interface McpError {
  code: McpErrorCode;
  message: string;
  data?: any;
}

export function createMcpError(
  code: McpErrorCode,
  message: string,
  data?: any
): McpError {
  return {
    code,
    message,
    data,
  };
}
