package integration

import (
	"context"
	"testing"

	"github.com/algonius/algonius-wallet/native/tests/integration/env"
	"github.com/stretchr/testify/require"
)

func TestGetBalanceTool(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	client := testEnv.GetMcpClient()
	require.NotNil(t, client, "MCP client should not be nil")

	require.NoError(t, client.Initialize(ctx), "failed to initialize MCP client")

	// Test valid request
	args := map[string]interface{}{
		"address": "0x1234567890123456789012345678901234567890",
		"token":   "ETH",
	}
	result, err := client.CallTool("get_balance", args)
	require.NoError(t, err, "failed to call get_balance tool")
	require.NotNil(t, result, "get_balance tool result should not be nil")

	// TODO: Add more specific assertions once we know the result structure
	// For now, just ensure the tool can be called successfully

	// Test missing address parameter
	args = map[string]interface{}{
		"token": "ETH",
	}
	result, err = client.CallTool("get_balance", args)
	// Should return an error for missing required parameter
	require.NoError(t, err) // The call itself should succeed
	require.NotNil(t, result)
	require.True(t, result.IsError, "should return error for missing address")

	// Test missing token parameter
	args = map[string]interface{}{
		"address": "0x1234567890123456789012345678901234567890",
	}
	result, err = client.CallTool("get_balance", args)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, result.IsError, "should return error for missing token")
}
