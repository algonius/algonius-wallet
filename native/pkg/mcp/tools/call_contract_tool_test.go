package tools

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCallContractToolHandlerSuccess(t *testing.T) {
	tool := NewCallContractTool()
	handler := tool.GetHandler()

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "call_contract",
			Arguments: map[string]any{
				"chain":            "ETH",
				"contract_address": "0x2222222222222222222222222222222222222222",
				"method":           "balanceOf",
				"args":             "[\"0x1111111111111111111111111111111111111111\"]",
			},
		},
	}

	result, err := handler(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, result.IsError)

	textContent, ok := mcp.AsTextContent(result.Content[0])
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "### Contract Call Result")
	assert.Contains(t, textContent.Text, "**Chain**: `ethereum`")
	assert.Contains(t, textContent.Text, "**Result**: `0`")
}

func TestCallContractToolHandlerRevert(t *testing.T) {
	tool := NewCallContractTool()
	handler := tool.GetHandler()

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "call_contract",
			Arguments: map[string]any{
				"chain":            "bsc",
				"contract_address": "0x2222222222222222222222222222222222222222",
				"method":           "revert",
			},
		},
	}

	result, err := handler(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsError)
}

func TestCallContractToolHandlerInvalidContractAddress(t *testing.T) {
	tool := NewCallContractTool()
	handler := tool.GetHandler()

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "call_contract",
			Arguments: map[string]any{
				"chain":            "solana",
				"contract_address": "0x-not-solana",
				"method":           "symbol",
			},
		},
	}

	result, err := handler(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsError)
}
