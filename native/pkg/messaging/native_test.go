package messaging

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/algonius/algonius-wallet/native/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func TestNewNativeMessaging_ValidConfig(t *testing.T) {
	log := logger.NewMockLogger()
	nm, err := NewNativeMessaging(NativeMessagingConfig{Logger: log})
	assert.NoError(t, err)
	assert.NotNil(t, nm)
	assert.Equal(t, log, nm.logger)
}

func TestNewNativeMessaging_NilLogger(t *testing.T) {
	nm, err := NewNativeMessaging(NativeMessagingConfig{})
	assert.Nil(t, nm)
	assert.Error(t, err)
}

func TestRegisterHandlerAndSendMessage(t *testing.T) {
	log := logger.NewMockLogger()
	var out bytes.Buffer
	nm, _ := NewNativeMessaging(NativeMessagingConfig{
		Logger: log,
		Stdout: &out,
	})
	received := make(chan Message, 1)
	nm.RegisterHandler("test", func(data interface{}) error {
		// In handleMessage, for unknown types, data is message.Data (json.RawMessage)
		msg, ok := data.(json.RawMessage)
		if !ok {
			// If not, try []byte
			b, ok2 := data.([]byte)
			assert.True(t, ok2)
			msg = json.RawMessage(b)
		}
		var val string
		_ = json.Unmarshal(msg, &val)
		assert.Equal(t, "hello", val)
		received <- Message{Type: "test", Data: msg}
		return nil
	})

	// Directly call handleMessage to simulate message delivery
	val, _ := json.Marshal("hello")
	msg := Message{Type: "test", Data: val}
	err := nm.handleMessage(msg)
	assert.NoError(t, err)

	select {
	case got := <-received:
		assert.Equal(t, "test", got.Type)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("handler not called")
	}
}

func TestSendMessage_WritesToStdout(t *testing.T) {
	log := logger.NewMockLogger()
	var out bytes.Buffer
	nm, _ := NewNativeMessaging(NativeMessagingConfig{
		Logger: log,
		Stdout: &out,
	})
	msg := Message{Type: "echo", Data: json.RawMessage(`"ping"`)}
	err := nm.SendMessage(msg)
	assert.NoError(t, err)
	// Check output format: 4 bytes length + JSON
	outBytes := out.Bytes()
	assert.GreaterOrEqual(t, len(outBytes), 5)
	length := binary.LittleEndian.Uint32(outBytes[:4])
	assert.Equal(t, uint32(len(outBytes[4:])), length)
	var got Message
	_ = json.Unmarshal(outBytes[4:], &got)
	assert.Equal(t, "echo", got.Type)
}

func TestRpcRequest_Success(t *testing.T) {
	log := logger.NewMockLogger()
	var out bytes.Buffer
	nm, _ := NewNativeMessaging(NativeMessagingConfig{
		Logger: log,
		Stdout: &out,
	})
	nm.RegisterRpcMethod("add", func(req RpcRequest) (RpcResponse, error) {
		var params []int
		_ = json.Unmarshal(req.Params, &params)
		sum := 0
		for _, v := range params {
			sum += v
		}
		result, _ := json.Marshal(sum)
		return RpcResponse{Result: result}, nil
	})

	params, _ := json.Marshal([]int{2, 3})
	respCh := make(chan RpcResponse, 1)
	var reqID string
	go func() {
		resp, err := nm.RpcRequest(RpcRequest{
			Method: "add",
			Params: params,
		}, RpcOptions{Timeout: 1000})
		assert.NoError(t, err)
		var sum int
		_ = json.Unmarshal(resp.Result, &sum)
		assert.Equal(t, 5, sum)
		respCh <- resp
	}()
	// Wait a moment for the request to be registered
	time.Sleep(10 * time.Millisecond)
	// Find the pending request ID
	nm.mutex.Lock()
	for id := range nm.pendingRequests {
		reqID = id
		break
	}
	nm.mutex.Unlock()
	// Simulate the response
	resp := Message{
		Type:   "rpc_response",
		ID:     reqID,
		Result: json.RawMessage(`5`),
	}
	_ = nm.handleMessage(resp)

	select {
	case <-respCh:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("RPC response not received")
	}
}

func TestRpcRequest_Timeout(t *testing.T) {
	log := logger.NewMockLogger()
	nm, _ := NewNativeMessaging(NativeMessagingConfig{
		Logger: log,
	})
	nm.RegisterRpcMethod("never", func(req RpcRequest) (RpcResponse, error) {
		time.Sleep(2 * time.Second)
		return RpcResponse{}, nil
	})
	params, _ := json.Marshal([]int{1})
	start := time.Now()
	_, err := nm.RpcRequest(RpcRequest{
		Method: "never",
		Params: params,
	}, RpcOptions{Timeout: 50})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
	assert.Less(t, time.Since(start).Milliseconds(), int64(500))
}

func TestHandleMessage_NoHandler(t *testing.T) {
	log := logger.NewMockLogger()
	var out bytes.Buffer
	nm, _ := NewNativeMessaging(NativeMessagingConfig{
		Logger: log,
		Stdout: &out,
	})
	msg := Message{Type: "unknown"}
	err := nm.handleMessage(msg)
	assert.NoError(t, err)
	assert.Contains(t, log.Infos, "Received message")
	// The error is sent as a message, not logged; check output buffer
	outBytes := out.Bytes()
	assert.GreaterOrEqual(t, len(outBytes), 5)
	length := binary.LittleEndian.Uint32(outBytes[:4])
	assert.Equal(t, uint32(len(outBytes[4:])), length)
	var got Message
	_ = json.Unmarshal(outBytes[4:], &got)
	assert.Equal(t, "error", got.Type)
	assert.Contains(t, got.Error.Message, "Unknown message type: unknown")
}

func TestProcessBuffer_InvalidJSON(t *testing.T) {
	log := logger.NewMockLogger()
	nm, _ := NewNativeMessaging(NativeMessagingConfig{
		Logger: log,
	})
	// Write invalid JSON message
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, uint32(6))
	buf.Write([]byte("{oops}"))
	nm.buffer = buf.Bytes()
	nm.processBuffer()
	assert.NotEmpty(t, log.Errors)
}

func TestRegisterRpcMethod_RegistersRequestHandler(t *testing.T) {
	log := logger.NewMockLogger()
	nm, _ := NewNativeMessaging(NativeMessagingConfig{
		Logger: log,
	})
	called := false
	nm.RegisterRpcMethod("foo", func(req RpcRequest) (RpcResponse, error) {
		called = true
		return RpcResponse{}, nil
	})
	handler, ok := nm.messageHandlers["rpc_request"]
	assert.True(t, ok)
	assert.NotNil(t, handler)
	// Simulate rpc_request message
	req := RpcRequest{Method: "foo"}
	err := handler(req)
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestSendMessage_ErrorMarshal(t *testing.T) {
	log := logger.NewMockLogger()
	nm, _ := NewNativeMessaging(NativeMessagingConfig{
		Logger: log,
		Stdout: io.Discard,
	})
	// Message with non-marshalable field
	msg := Message{Type: "bad", Data: json.RawMessage{0xff, 0xfe, 0xfd}}
	// Intentionally break JSON
	nm.mutex.Lock()
	nm.stdout = &brokenWriter{}
	nm.mutex.Unlock()
	err := nm.SendMessage(msg)
	assert.Error(t, err)
}

type brokenWriter struct{}

func (w *brokenWriter) Write(p []byte) (int, error) {
	return 0, errors.New("write error")
}
