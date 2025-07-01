// Package messaging implements Chrome Native Messaging protocol for Algonius Native Host.
package messaging

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/algonius/algonius-wallet/native/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// NativeMessaging implements Chrome native messaging protocol.
type NativeMessaging struct {
	logger          logger.Logger
	stdin           io.Reader
	stdout          io.Writer
	buffer          []byte
	messageHandlers map[string]MessageHandler
	rpcHandlers     map[string]RpcHandler
	pendingRequests map[string]*pendingRequest
	mutex           sync.Mutex
}

type pendingRequest struct {
	done     chan struct{}
	response RpcResponse
	err      error
	timer    *time.Timer
}

// NativeMessagingConfig configures NativeMessaging.
type NativeMessagingConfig struct {
	Logger logger.Logger
	Stdin  io.Reader
	Stdout io.Writer
}

// NewNativeMessaging creates a new NativeMessaging instance.
func NewNativeMessaging(config NativeMessagingConfig) (*NativeMessaging, error) {
	if config.Logger == nil {
		return nil, fmt.Errorf("logger is required")
	}
	stdin := config.Stdin
	if stdin == nil {
		stdin = os.Stdin
	}
	stdout := config.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	nm := &NativeMessaging{
		logger:          config.Logger,
		stdin:           stdin,
		stdout:          stdout,
		buffer:          make([]byte, 0),
		messageHandlers: make(map[string]MessageHandler),
		rpcHandlers:     make(map[string]RpcHandler),
		pendingRequests: make(map[string]*pendingRequest),
	}
	nm.registerRpcResponseHandler()
	return nm, nil
}

// Start begins processing messages from stdin.
func (nm *NativeMessaging) Start() error {
	nm.logger.Info("Starting native messaging processing")
	go func() {
		buffer := make([]byte, 4096)
		for {
			n, err := nm.stdin.Read(buffer)
			if err != nil {
				if err == io.EOF {
					nm.logger.Info("Native messaging: stdin closed")
					return
				}
				nm.logger.Error("Error reading from stdin", zap.Error(err))
				return
			}
			if n > 0 {
				nm.buffer = append(nm.buffer, buffer[:n]...)
				nm.processBuffer()
			}
		}
	}()
	return nil
}

// processBuffer processes the buffer for messages.
func (nm *NativeMessaging) processBuffer() {
	for len(nm.buffer) >= 4 {
		messageLength := binary.LittleEndian.Uint32(nm.buffer[:4])
		if uint32(len(nm.buffer)) < messageLength+4 {
			return
		}
		messageJSON := nm.buffer[4 : messageLength+4]
		nm.buffer = nm.buffer[messageLength+4:]

		var message Message
		if err := json.Unmarshal(messageJSON, &message); err != nil {
			nm.logger.Error("Error parsing message JSON", zap.Error(err), zap.String("json", string(messageJSON)))
			continue
		}
		go func(msg Message) {
			if err := nm.handleMessage(msg); err != nil {
				nm.logger.Error("Error handling message", zap.Error(err), zap.Any("message", msg))
			}
		}(message)
	}
}

// handleMessage processes a received message.
func (nm *NativeMessaging) handleMessage(message Message) error {
	nm.logger.Info("Received message", zap.Any("message", message))
	handler, ok := nm.messageHandlers[message.Type]
	if !ok {
		nm.logger.Warn("No handler registered for message type", zap.String("type", message.Type))
		return nm.SendMessage(Message{
			Type:  "error",
			Error: &ErrorInfo{Message: fmt.Sprintf("Unknown message type: %s", message.Type)},
		})
	}
	var data interface{}
	switch message.Type {
	case "rpc_request":
		data = RpcRequest{
			ID:     message.ID,
			Method: message.Method,
			Params: message.Params,
		}
	case "rpc_response":
		data = RpcResponse{
			ID:     message.ID,
			Result: message.Result,
			Error:  message.Error,
		}
	default:
		data = message.Data
	}
	return handler(data)
}

// RegisterHandler registers a handler for a specific message type.
func (nm *NativeMessaging) RegisterHandler(messageType string, handler MessageHandler) {
	nm.logger.Debug("Registering handler for message type", zap.String("type", messageType))
	nm.messageHandlers[messageType] = handler
}

// RegisterRpcMethod registers a handler for an RPC method.
func (nm *NativeMessaging) RegisterRpcMethod(method string, handler RpcHandler) {
	nm.logger.Debug("Registering RPC handler for method", zap.String("method", method))
	nm.rpcHandlers[method] = handler
	if _, exists := nm.messageHandlers["rpc_request"]; !exists {
		nm.registerRpcRequestHandler()
	}
}

// SendMessage sends a message to stdout.
func (nm *NativeMessaging) SendMessage(message Message) error {
	nm.logger.Debug("Sending message", zap.Any("message", message))
	messageJSON, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("error marshaling message: %w", err)
	}
	messageLength := uint32(len(messageJSON))
	buffer := make([]byte, 4+messageLength)
	binary.LittleEndian.PutUint32(buffer, messageLength)
	copy(buffer[4:], messageJSON)
	nm.mutex.Lock()
	defer nm.mutex.Unlock()
	_, err = nm.stdout.Write(buffer)
	if err != nil {
		return fmt.Errorf("error writing message: %w", err)
	}
	return nil
}

// RpcRequest sends an RPC request and waits for the response.
func (nm *NativeMessaging) RpcRequest(request RpcRequest, options RpcOptions) (RpcResponse, error) {
	id := request.ID
	if id == "" {
		id = uuid.New().String()
		request.ID = id
	}
	nm.logger.Info("Sending RPC request", zap.String("method", request.Method), zap.String("id", id))
	timeout := 5000
	if options.Timeout > 0 {
		timeout = options.Timeout
	}
	pending := &pendingRequest{
		done: make(chan struct{}),
	}
	pending.timer = time.AfterFunc(time.Duration(timeout)*time.Millisecond, func() {
		nm.mutex.Lock()
		defer nm.mutex.Unlock()
		if _, exists := nm.pendingRequests[id]; exists {
			pending.err = fmt.Errorf("RPC request timeout: %s (id: %s)", request.Method, id)
			close(pending.done)
			delete(nm.pendingRequests, id)
		}
	})
	nm.mutex.Lock()
	nm.pendingRequests[id] = pending
	nm.mutex.Unlock()
	if err := nm.SendMessage(Message{
		Type:   "rpc_request",
		ID:     id,
		Method: request.Method,
		Params: request.Params,
	}); err != nil {
		nm.mutex.Lock()
		delete(nm.pendingRequests, id)
		nm.mutex.Unlock()
		pending.timer.Stop()
		return RpcResponse{}, err
	}
	<-pending.done
	nm.logger.Info("RPC request Done", zap.Any("response", pending.response), zap.Any("err", pending.err))
	return pending.response, pending.err
}

// registerRpcResponseHandler registers a handler for RPC responses.
func (nm *NativeMessaging) registerRpcResponseHandler() {
	nm.RegisterHandler("rpc_response", func(data interface{}) error {
		response, ok := data.(RpcResponse)
		if !ok {
			return fmt.Errorf("invalid RPC response format")
		}
		id := response.ID
		nm.logger.Debug("Received RPC response for ID", zap.String("id", id))
		nm.mutex.Lock()
		defer nm.mutex.Unlock()
		pending, exists := nm.pendingRequests[id]
		if !exists {
			nm.logger.Warn("No pending request found for RPC response ID", zap.String("id", id))
			return nil
		}
		pending.timer.Stop()
		pending.response = response
		close(pending.done)
		delete(nm.pendingRequests, id)
		return nil
	})
}

// registerRpcRequestHandler registers a handler for RPC requests.
func (nm *NativeMessaging) registerRpcRequestHandler() {
	nm.RegisterHandler("rpc_request", func(data interface{}) error {
		request, ok := data.(RpcRequest)
		if !ok {
			return fmt.Errorf("invalid RPC request format")
		}
		method := request.Method
		nm.logger.Debug("Handling RPC request", zap.String("method", method))
		handler, exists := nm.rpcHandlers[method]
		if !exists {
			nm.logger.Warn("No handler registered for RPC method", zap.String("method", method))
			return nm.SendMessage(Message{
				Type: "rpc_response",
				ID:   request.ID,
				Error: &ErrorInfo{
					Code:    -32601,
					Message: fmt.Sprintf("Method not found: %s", method),
				},
			})
		}
		response, err := handler(request)
		if err != nil {
			nm.logger.Error("Error in RPC handler", zap.Error(err))
			return nm.SendMessage(Message{
				Type: "rpc_response",
				ID:   request.ID,
				Error: &ErrorInfo{
					Code:    -32000,
					Message: fmt.Sprintf("Server error: %s", err.Error()),
				},
			})
		}
		response.ID = request.ID
		return nm.SendMessage(Message{
			Type:   "rpc_response",
			ID:     response.ID,
			Result: response.Result,
			Error:  response.Error,
		})
	})
}
