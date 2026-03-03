package tools

import (
	"context"
	"testing"

	"github.com/algonius/algonius-wallet/native/pkg/wallet/chain"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockChainForEstimateGas struct {
	estimateFail bool
}

func (m *mockChainForEstimateGas) CreateWallet(ctx context.Context) (*chain.WalletInfo, error) {
	return nil, nil
}

func (m *mockChainForEstimateGas) ImportFromMnemonic(ctx context.Context, mnemonic, derivationPath string) (*chain.WalletInfo, error) {
	return nil, nil
}

func (m *mockChainForEstimateGas) GetBalance(ctx context.Context, address string, token string) (string, error) {
	return "", nil
}

func (m *mockChainForEstimateGas) SendTransaction(ctx context.Context, from, to, amount, token, privateKey string) (string, error) {
	return "", nil
}

func (m *mockChainForEstimateGas) EstimateGas(ctx context.Context, from, to, amount, token string) (uint64, string, error) {
	if m.estimateFail {
		return 0, "", assert.AnError
	}
	return 52000, "15", nil
}

func (m *mockChainForEstimateGas) ConfirmTransaction(ctx context.Context, txHash string, requiredConfirmations uint64) (*chain.TransactionConfirmation, error) {
	return nil, nil
}

func (m *mockChainForEstimateGas) SignMessage(privateKeyHex, message string) (string, error) {
	return "", nil
}

func (m *mockChainForEstimateGas) GetChainName() string {
	return "mock"
}

func TestEstimateGasToolHandlerSuccess(t *testing.T) {
	tool := NewEstimateGasTool(nil)
	tool.getChainInterface = func(chainName string) (chain.IChain, error) {
		return &mockChainForEstimateGas{}, nil
	}
	handler := tool.GetHandler()

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "estimate_gas",
			Arguments: map[string]any{
				"chain":  "ETH",
				"from":   "0x1111111111111111111111111111111111111111",
				"to":     "0x2222222222222222222222222222222222222222",
				"amount": "1.2",
			},
		},
	}

	result, err := handler(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, result.IsError)

	textContent, ok := mcp.AsTextContent(result.Content[0])
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "### Gas Estimate")
	assert.Contains(t, textContent.Text, "**Chain**: `ethereum`")
	assert.Contains(t, textContent.Text, "**Gas Limit**: `52000`")
	assert.Contains(t, textContent.Text, "**Gas Price**: `15`")
}

func TestEstimateGasToolHandlerMissingAmount(t *testing.T) {
	tool := NewEstimateGasTool(nil)
	handler := tool.GetHandler()

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "estimate_gas",
			Arguments: map[string]any{
				"chain": "bsc",
				"from":  "0x1111111111111111111111111111111111111111",
				"to":    "0x2222222222222222222222222222222222222222",
			},
		},
	}

	result, err := handler(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsError)
}

func TestEstimateGasToolHandlerEstimateError(t *testing.T) {
	tool := NewEstimateGasTool(nil)
	tool.getChainInterface = func(chainName string) (chain.IChain, error) {
		return &mockChainForEstimateGas{estimateFail: true}, nil
	}
	handler := tool.GetHandler()

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "estimate_gas",
			Arguments: map[string]any{
				"chain":  "solana",
				"from":   "FnVyf9f7hFmA6N5HtV6nQWmvMRGsiE9zraFMvx6bMpiK",
				"to":     "5oNDL3swdJJF1g9DzJiZ4ynHXgszjAEpUkxVYejchzrY",
				"amount": "0.5",
			},
		},
	}

	result, err := handler(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsError)
}
