// Package messaging defines types for native messaging and RPC.
package messaging

import (
	"encoding/json"
)

// Message represents a generic message for native messaging.
type Message struct {
	Type   string          `json:"type"`
	ID     string          `json:"id,omitempty"`
	Method string          `json:"method,omitempty"`
	Params json.RawMessage `json:"params,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  *ErrorInfo      `json:"error,omitempty"`
	Data   json.RawMessage `json:"data,omitempty"`
}

// ErrorInfo represents error details in a message.
type ErrorInfo struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message"`
}

// RpcRequest represents an RPC request message.
type RpcRequest struct {
	ID     string          `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params,omitempty"`
}

// RpcResponse represents an RPC response message.
type RpcResponse struct {
	ID     string          `json:"id"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  *ErrorInfo      `json:"error,omitempty"`
}

// RpcOptions configures RPC request behavior.
type RpcOptions struct {
	Timeout int // milliseconds
}

// MessageHandler handles a generic message or data.
type MessageHandler func(data interface{}) error

// RpcHandler handles an RPC request and returns a response.
type RpcHandler func(request RpcRequest) (RpcResponse, error)
