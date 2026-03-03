package tools

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeployContractToolHandlerSuccessEVM(t *testing.T) {
	tool := NewDeployContractTool()
	handler := tool.GetHandler()

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "deploy_contract",
			Arguments: map[string]any{
				"chain":            "bsc",
				"from":             "0x1111111111111111111111111111111111111111",
				"bytecode":         "0x6080604052348015600f57600080fd5b5060",
				"constructor_args": "[]",
			},
		},
	}

	result, err := handler(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, result.IsError)

	textContent, ok := mcp.AsTextContent(result.Content[0])
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "### Contract Deployment Submitted")
	assert.Contains(t, textContent.Text, "**Chain**: `bsc`")
	assert.Contains(t, textContent.Text, "**Status**: `pending`")
}

func TestDeployContractToolHandlerSuccessSolana(t *testing.T) {
	tool := NewDeployContractTool()
	handler := tool.GetHandler()

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "deploy_contract",
			Arguments: map[string]any{
				"chain":    "sol",
				"from":     "FnVyf9f7hFmA6N5HtV6nQWmvMRGsiE9zraFMvx6bMpiK",
				"bytecode": "solana_program_blob",
			},
		},
	}

	result, err := handler(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, result.IsError)

	textContent, ok := mcp.AsTextContent(result.Content[0])
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "**Chain**: `solana`")
	assert.Contains(t, textContent.Text, "**Contract Address**:")
}

func TestDeployContractToolHandlerInvalidAddress(t *testing.T) {
	tool := NewDeployContractTool()
	handler := tool.GetHandler()

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "deploy_contract",
			Arguments: map[string]any{
				"chain":    "ethereum",
				"from":     "invalid-address",
				"bytecode": "0x60806040",
			},
		},
	}

	result, err := handler(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsError)
}

func TestDeployContractToolHandlerInvalidBytecode(t *testing.T) {
	tool := NewDeployContractTool()
	handler := tool.GetHandler()

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "deploy_contract",
			Arguments: map[string]any{
				"chain":    "ethereum",
				"from":     "0x1111111111111111111111111111111111111111",
				"bytecode": "not-hex-bytecode",
			},
		},
	}

	result, err := handler(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsError)
}
