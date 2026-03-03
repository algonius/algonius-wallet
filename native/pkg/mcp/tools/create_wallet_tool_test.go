package tools

import (
	"context"
	"testing"

	"github.com/algonius/algonius-wallet/native/pkg/wallet"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockWalletManagerForCreateWallet struct {
	*wallet.MockWalletManager
	lastChain    string
	lastPassword string
	shouldFail   bool
}

func (m *mockWalletManagerForCreateWallet) CreateWallet(ctx context.Context, chain, password string) (string, string, string, error) {
	m.lastChain = chain
	m.lastPassword = password
	if m.shouldFail {
		return "", "", "", assert.AnError
	}
	return "0x1111111111111111111111111111111111111111", "0xpublic", "mnemonic phrase", nil
}

func TestCreateWalletToolMeta(t *testing.T) {
	tool := NewCreateWalletTool(&wallet.MockWalletManager{})
	meta := tool.GetMeta()

	assert.Equal(t, "create_wallet", meta.Name)
	assert.Contains(t, meta.InputSchema.Properties, "chain")
	assert.Contains(t, meta.InputSchema.Required, "chain")
}

func TestCreateWalletToolHandlerSuccess(t *testing.T) {
	mockManager := &mockWalletManagerForCreateWallet{MockWalletManager: &wallet.MockWalletManager{}}
	tool := NewCreateWalletTool(mockManager)
	handler := tool.GetHandler()

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "create_wallet",
			Arguments: map[string]any{
				"chain": "ETH",
			},
		},
	}

	result, err := handler(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, result.IsError)
	assert.Equal(t, "ethereum", mockManager.lastChain)
	assert.Equal(t, "temp-mcp-password-123", mockManager.lastPassword)

	textContent, ok := mcp.AsTextContent(result.Content[0])
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "### Wallet Created")
	assert.Contains(t, textContent.Text, "**Chain**: `ethereum`")
	assert.Contains(t, textContent.Text, "0x1111111111111111111111111111111111111111")
}

func TestCreateWalletToolHandlerMissingChain(t *testing.T) {
	tool := NewCreateWalletTool(&wallet.MockWalletManager{})
	handler := tool.GetHandler()

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "create_wallet",
			Arguments: map[string]any{},
		},
	}

	result, err := handler(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsError)
}

func TestCreateWalletToolHandlerManagerError(t *testing.T) {
	mockManager := &mockWalletManagerForCreateWallet{
		MockWalletManager: &wallet.MockWalletManager{},
		shouldFail:        true,
	}
	tool := NewCreateWalletTool(mockManager)
	handler := tool.GetHandler()

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "create_wallet",
			Arguments: map[string]any{
				"chain": "solana",
			},
		},
	}

	result, err := handler(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsError)
}
