package tools

import (
	"context"
	"testing"

	"github.com/algonius/algonius-wallet/native/pkg/wallet"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockWalletManagerForGetBalance struct {
	*wallet.MockWalletManager
	lastAddress string
	lastToken   string
	shouldFail  bool
}

func (m *mockWalletManagerForGetBalance) GetBalance(ctx context.Context, address, token string) (string, error) {
	m.lastAddress = address
	m.lastToken = token
	if m.shouldFail {
		return "", assert.AnError
	}
	return "123.456", nil
}

func TestGetBalanceToolHandlerSuccess(t *testing.T) {
	mockManager := &mockWalletManagerForGetBalance{MockWalletManager: &wallet.MockWalletManager{}}
	tool := NewGetBalanceTool(mockManager)
	handler := tool.GetHandler()

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "get_balance",
			Arguments: map[string]any{
				"address": "0x742d35Cc6634C0532925a3b8D4C2B79C2b86A7A8",
				"token":   "ETH",
			},
		},
	}

	result, err := handler(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, result.IsError)
	assert.Equal(t, "0x742d35Cc6634C0532925a3b8D4C2B79C2b86A7A8", mockManager.lastAddress)
	assert.Equal(t, "ETH", mockManager.lastToken)

	textContent, ok := mcp.AsTextContent(result.Content[0])
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "### Wallet Balance")
	assert.Contains(t, textContent.Text, "123.456")
}

func TestGetBalanceToolHandlerMissingAddress(t *testing.T) {
	tool := NewGetBalanceTool(&wallet.MockWalletManager{})
	handler := tool.GetHandler()

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "get_balance",
			Arguments: map[string]any{
				"token": "ETH",
			},
		},
	}

	result, err := handler(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsError)
}

func TestGetBalanceToolHandlerManagerError(t *testing.T) {
	mockManager := &mockWalletManagerForGetBalance{
		MockWalletManager: &wallet.MockWalletManager{},
		shouldFail:        true,
	}
	tool := NewGetBalanceTool(mockManager)
	handler := tool.GetHandler()

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "get_balance",
			Arguments: map[string]any{
				"address": "0x742d35Cc6634C0532925a3b8D4C2B79C2b86A7A8",
				"token":   "ETH",
			},
		},
	}

	result, err := handler(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsError)
}
