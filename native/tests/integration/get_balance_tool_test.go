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

	// Test BSC native token (BNB) - should work with BSC addresses
	t.Run("BSC_BNB_Balance", func(t *testing.T) {
		args := map[string]interface{}{
			"address": "0x1234567890123456789012345678901234567890", // BSC address (same format as ETH)
			"token":   "BNB",
		}
		result, err := client.CallTool("get_balance", args)
		require.NoError(t, err, "failed to call get_balance tool for BNB")
		require.NotNil(t, result, "get_balance tool result should not be nil for BNB")
		
		// Log result for debugging
		t.Logf("BNB Balance Result: IsError=%v, Content=%+v", result.IsError, result.Content)
		
		require.False(t, result.IsError, "should successfully query BNB balance")
	})

	// Test Solana native token (SOL) - should work with Solana addresses  
	t.Run("Solana_SOL_Balance", func(t *testing.T) {
		args := map[string]interface{}{
			"address": "11111111111111111111111111111112", // Valid Solana address format
			"token":   "SOL",
		}
		result, err := client.CallTool("get_balance", args)
		require.NoError(t, err, "failed to call get_balance tool for SOL")
		require.NotNil(t, result, "get_balance tool result should not be nil for SOL")
		
		// Log result for debugging
		t.Logf("SOL Balance Result: IsError=%v, Content=%+v", result.IsError, result.Content)
		
		require.False(t, result.IsError, "should successfully query SOL balance")
	})
	
	// Test case that should fail: unsupported token on wrong chain
	t.Run("Unsupported_Token_Should_Fail", func(t *testing.T) {
		args := map[string]interface{}{
			"address": "0x1234567890123456789012345678901234567890",
			"token":   "UNSUPPORTED_TOKEN",
		}
		result, err := client.CallTool("get_balance", args)
		require.NoError(t, err, "call should succeed but return error result")
		require.NotNil(t, result, "result should not be nil")
		
		// Log result for debugging  
		t.Logf("Unsupported Token Result: IsError=%v, Content=%+v", result.IsError, result.Content)
		
		require.True(t, result.IsError, "should return error for unsupported token")
	})

	// Test cross-chain token standardization - BNB with Ethereum address should work
	t.Run("Cross_Chain_BNB_Query", func(t *testing.T) {
		args := map[string]interface{}{
			"address": "0x1234567890123456789012345678901234567890",
			"token":   "BNB",
		}
		result, err := client.CallTool("get_balance", args)
		require.NoError(t, err, "failed to call get_balance tool for cross-chain BNB")
		require.NotNil(t, result, "get_balance tool result should not be nil for cross-chain BNB")
		require.False(t, result.IsError, "should successfully query BNB balance on BSC chain")
	})

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
