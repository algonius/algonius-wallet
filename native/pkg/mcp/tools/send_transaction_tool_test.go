package tools

import (
	"context"
	"testing"

	"github.com/algonius/algonius-wallet/native/pkg/wallet"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockWalletManagerForSendTransaction struct {
	*wallet.MockWalletManager
	lastEstimateChain string
	lastSendChain     string
	estimateFail      bool
	sendFail          bool
}

func (m *mockWalletManagerForSendTransaction) EstimateGas(ctx context.Context, chain, from, to, amount, token string) (uint64, string, error) {
	m.lastEstimateChain = chain
	if m.estimateFail {
		return 0, "", assert.AnError
	}
	return 21000, "20", nil
}

func (m *mockWalletManagerForSendTransaction) SendTransaction(ctx context.Context, chain, from, to, amount, token string) (string, error) {
	m.lastSendChain = chain
	if m.sendFail {
		return "", assert.AnError
	}
	return "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", nil
}

func TestSendTransactionToolHandlerSuccess(t *testing.T) {
	mockManager := &mockWalletManagerForSendTransaction{MockWalletManager: &wallet.MockWalletManager{}}
	tool := NewSendTransactionTool(mockManager)
	handler := tool.GetHandler()

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "send_transaction",
			Arguments: map[string]any{
				"chain":  "SOL",
				"from":   "FnVyf9f7hFmA6N5HtV6nQWmvMRGsiE9zraFMvx6bMpiK",
				"to":     "5oNDL3swdJJF1g9DzJiZ4ynHXgszjAEpUkxVYejchzrY",
				"amount": "0.2",
				"token":  "SOL",
			},
		},
	}

	result, err := handler(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, result.IsError)
	assert.Equal(t, "solana", mockManager.lastEstimateChain)
	assert.Equal(t, "solana", mockManager.lastSendChain)

	textContent, ok := mcp.AsTextContent(result.Content[0])
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "### Transaction Sent")
	assert.Contains(t, textContent.Text, "**Chain**: `solana`")
	assert.Contains(t, textContent.Text, "**Gas Limit**: `21000`")
}

func TestSendTransactionToolHandlerInvalidChain(t *testing.T) {
	tool := NewSendTransactionTool(&wallet.MockWalletManager{})
	handler := tool.GetHandler()

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "send_transaction",
			Arguments: map[string]any{
				"chain":  "unknown-chain",
				"from":   "0x1111111111111111111111111111111111111111",
				"to":     "0x2222222222222222222222222222222222222222",
				"amount": "1",
			},
		},
	}

	result, err := handler(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsError)
}

func TestSendTransactionToolHandlerSendError(t *testing.T) {
	mockManager := &mockWalletManagerForSendTransaction{
		MockWalletManager: &wallet.MockWalletManager{},
		sendFail:          true,
	}
	tool := NewSendTransactionTool(mockManager)
	handler := tool.GetHandler()

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "send_transaction",
			Arguments: map[string]any{
				"chain":  "ethereum",
				"from":   "0x1111111111111111111111111111111111111111",
				"to":     "0x2222222222222222222222222222222222222222",
				"amount": "1",
			},
		},
	}

	result, err := handler(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsError)
}
