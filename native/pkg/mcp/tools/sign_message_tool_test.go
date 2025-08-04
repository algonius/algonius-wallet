package tools

import (
	"context"
	"testing"

	"github.com/algonius/algonius-wallet/native/pkg/wallet"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockWalletManagerWithSignMessage is a mock implementation of wallet.IWalletManager for testing sign_message tool
type MockWalletManagerWithSignMessage struct {
	*wallet.MockWalletManager
	shouldReturnError bool
	mockSignature     string
}

func (m *MockWalletManagerWithSignMessage) SignMessage(ctx context.Context, address, message string) (string, error) {
	if m.shouldReturnError {
		return "", assert.AnError
	}
	return m.mockSignature, nil
}

func TestSignMessageToolMeta(t *testing.T) {
	mockManager := &wallet.MockWalletManager{}
	tool := NewSignMessageTool(mockManager, nil)

	meta := tool.GetMeta()

	assert.Equal(t, "sign_message", meta.Name)
	assert.Equal(t, "Sign a text message or raw bytes with a wallet's private key", meta.Description)

	// Check that all expected parameters are present
	assert.Contains(t, meta.InputSchema.Properties, "address")
	assert.Contains(t, meta.InputSchema.Properties, "message")

	// Check that both parameters are required
	required := meta.InputSchema.Required
	assert.Contains(t, required, "address")
	assert.Contains(t, required, "message")
}

func TestSignMessageToolCreation(t *testing.T) {
	mockManager := &wallet.MockWalletManager{}
	tool := NewSignMessageTool(mockManager, nil)

	require.NotNil(t, tool)
	require.NotNil(t, tool.manager)

	// Test that the tool has the expected methods
	meta := tool.GetMeta()
	require.NotNil(t, meta)

	handler := tool.GetHandler()
	require.NotNil(t, handler)
}

func TestSignMessageToolHandler_Success(t *testing.T) {
	mockManager := &MockWalletManagerWithSignMessage{
		MockWalletManager: &wallet.MockWalletManager{},
		mockSignature:     "mock_signature_1234567890",
	}
	tool := NewSignMessageTool(mockManager, nil)
	handler := tool.GetHandler()

	// Create proper request structure
	params := mcp.CallToolParams{
		Name: "sign_message",
		Arguments: map[string]interface{}{
			"address": "0x742d35Cc6634C0532925a3b8D4C2B79C2b86A7A8",
			"message": "Hello, World!",
		},
	}

	req := mcp.CallToolRequest{}
	req.Params = params

	// Execute the handler
	result, err := handler(context.Background(), req)

	// Verify the result
	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, result.IsError)
	require.NotNil(t, result.Content)
	require.Len(t, result.Content, 1)

	// Check that the markdown output contains expected content
	textContent, ok := mcp.AsTextContent(result.Content[0])
	require.True(t, ok, "result should contain text content")
	markdown := textContent.Text
	assert.Contains(t, markdown, "### Message Signed Successfully")
	assert.Contains(t, markdown, "0x742d35Cc6634C0532925a3b8D4C2B79C2b86A7A8")
	assert.Contains(t, markdown, "ethereum")
	assert.Contains(t, markdown, "13 characters")  // Length of "Hello, World!"
	assert.Contains(t, markdown, "mock_signature_1234567890")
}

func TestSignMessageToolHandler_MissingAddress(t *testing.T) {
	mockManager := &wallet.MockWalletManager{}
	tool := NewSignMessageTool(mockManager, nil)
	handler := tool.GetHandler()

	// Create request without required address parameter
	params := mcp.CallToolParams{
		Name: "sign_message",
		Arguments: map[string]interface{}{
			"message": "Hello, World!",
		},
	}

	req := mcp.CallToolRequest{}
	req.Params = params

	// Execute the handler
	result, err := handler(context.Background(), req)

	// Verify that it returns an error
	require.NoError(t, err) // Handler shouldn't return Go error
	require.NotNil(t, result)
	require.True(t, result.IsError)
}

func TestSignMessageToolHandler_MissingMessage(t *testing.T) {
	mockManager := &wallet.MockWalletManager{}
	tool := NewSignMessageTool(mockManager, nil)
	handler := tool.GetHandler()

	// Create request without required message parameter
	params := mcp.CallToolParams{
		Name: "sign_message",
		Arguments: map[string]interface{}{
			"address": "0x742d35Cc6634C0532925a3b8D4C2B79C2b86A7A8",
		},
	}

	req := mcp.CallToolRequest{}
	req.Params = params

	// Execute the handler
	result, err := handler(context.Background(), req)

	// Verify that it returns an error
	require.NoError(t, err) // Handler shouldn't return Go error
	require.NotNil(t, result)
	require.True(t, result.IsError)
}

func TestSignMessageToolHandler_EmptyAddress(t *testing.T) {
	mockManager := &wallet.MockWalletManager{}
	tool := NewSignMessageTool(mockManager, nil)
	handler := tool.GetHandler()

	// Create request with empty address
	params := mcp.CallToolParams{
		Name: "sign_message",
		Arguments: map[string]interface{}{
			"address": "",
			"message": "Hello, World!",
		},
	}

	req := mcp.CallToolRequest{}
	req.Params = params

	// Execute the handler
	result, err := handler(context.Background(), req)

	// Verify that it returns an error
	require.NoError(t, err) // Handler shouldn't return Go error
	require.NotNil(t, result)
	require.True(t, result.IsError)
}

func TestSignMessageToolHandler_EmptyMessage(t *testing.T) {
	mockManager := &wallet.MockWalletManager{}
	tool := NewSignMessageTool(mockManager, nil)
	handler := tool.GetHandler()

	// Create request with empty message
	params := mcp.CallToolParams{
		Name: "sign_message",
		Arguments: map[string]interface{}{
			"address": "0x742d35Cc6634C0532925a3b8D4C2B79C2b86A7A8",
			"message": "",
		},
	}

	req := mcp.CallToolRequest{}
	req.Params = params

	// Execute the handler
	result, err := handler(context.Background(), req)

	// Verify that it returns an error
	require.NoError(t, err) // Handler shouldn't return Go error
	require.NotNil(t, result)
	require.True(t, result.IsError)
}

func TestSignMessageToolHandler_ManagerError(t *testing.T) {
	mockManager := &MockWalletManagerWithSignMessage{
		MockWalletManager: &wallet.MockWalletManager{},
		shouldReturnError: true,
	}
	tool := NewSignMessageTool(mockManager, nil)
	handler := tool.GetHandler()

	// Create request with required parameters
	params := mcp.CallToolParams{
		Name: "sign_message",
		Arguments: map[string]interface{}{
			"address": "0x742d35Cc6634C0532925a3b8D4C2B79C2b86A7A8",
			"message": "Hello, World!",
		},
	}

	req := mcp.CallToolRequest{}
	req.Params = params

	// Execute the handler
	result, err := handler(context.Background(), req)

	// Verify that it returns an error response
	require.NoError(t, err) // Handler shouldn't return Go error
	require.NotNil(t, result)
	require.True(t, result.IsError)
}

func TestSignMessageToolHandler_SolanaRawBytes(t *testing.T) {
	mockManager := &MockWalletManagerWithSignMessage{
		MockWalletManager: &wallet.MockWalletManager{},
		mockSignature:     "mock_signature_1234567890",
	}
	tool := NewSignMessageTool(mockManager, nil)
	handler := tool.GetHandler()

	// Create request with Solana raw bytes message
	params := mcp.CallToolParams{
		Name: "sign_message",
		Arguments: map[string]interface{}{
			"address": "solana_address_base58",
			"message": "__SOLANA_RAW_BYTES__:SGVsbG8sIFdvcmxkIQ==",
		},
	}

	req := mcp.CallToolRequest{}
	req.Params = params

	// Execute the handler
	result, err := handler(context.Background(), req)

	// Verify the result
	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, result.IsError)
	require.NotNil(t, result.Content)
	require.Len(t, result.Content, 1)

	// Check that the markdown output contains expected content for Solana
	textContent, ok := mcp.AsTextContent(result.Content[0])
	require.True(t, ok, "result should contain text content")
	markdown := textContent.Text
	assert.Contains(t, markdown, "### Message Signed Successfully")
	assert.Contains(t, markdown, "solana_address_base58")
	assert.Contains(t, markdown, "solana")
	assert.Contains(t, markdown, "41 characters")  // Length of "__SOLANA_RAW_BYTES__:SGVsbG8sIFdvcmxkIQ=="
	assert.Contains(t, markdown, "mock_signature_1234567890")
}