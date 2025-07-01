package integration

import (
	"context"
	"testing"

	"github.com/algonius/algonius-wallet/native/tests/integration/env"
	"github.com/stretchr/testify/require"
)

func TestSwapTokensTool(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	client := testEnv.GetMcpClient()
	require.NotNil(t, client, "MCP client should not be nil")

	require.NoError(t, client.Initialize(ctx), "failed to initialize MCP client")

	// Test exact input swap
	args := map[string]interface{}{
		"chain":              "ethereum",
		"token_in":           "0xA0b86a33E6441e064c7d56Fb65dd0E8F8b764C4e", // USDC
		"token_out":          "0xC02aaA39b223FE65608aaC29B7b0Ed646Cf8F7D5", // WETH
		"amount_in":          "1000.0",
		"from":               "0x1234567890123456789012345678901234567890",
		"slippage_tolerance": 0.5,
		"deadline":           300,
		"dex":                "uniswap",
	}
	result, err := client.CallTool("swap_tokens", args)
	require.NoError(t, err, "failed to call swap_tokens tool")
	require.NotNil(t, result, "swap_tokens tool result should not be nil")

	// Test exact output swap
	outputArgs := map[string]interface{}{
		"chain":              "ethereum",
		"token_in":           "0xC02aaA39b223FE65608aaC29B7b0Ed646Cf8F7D5", // WETH
		"token_out":          "0xA0b86a33E6441e064c7d56Fb65dd0E8F8b764C4e", // USDC
		"amount_out":         "1000.0",
		"from":               "0x1234567890123456789012345678901234567890",
		"slippage_tolerance": 1.0,
		"dex":                "uniswap",
	}
	outputResult, err := client.CallTool("swap_tokens", outputArgs)
	require.NoError(t, err, "failed to call swap_tokens tool with exact output")
	require.NotNil(t, outputResult, "swap_tokens tool result should not be nil for exact output")
}

func TestSwapTokensToolValidation(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	client := testEnv.GetMcpClient()
	require.NotNil(t, client, "MCP client should not be nil")

	require.NoError(t, client.Initialize(ctx), "failed to initialize MCP client")

	testCases := []struct {
		name string
		args map[string]interface{}
	}{
		{
			name: "missing_chain",
			args: map[string]interface{}{
				"token_in":  "0xA0b86a33E6441e064c7d56Fb65dd0E8F8b764C4e",
				"token_out": "0xC02aaA39b223FE65608aaC29B7b0Ed646Cf8F7D5",
				"amount_in": "1000.0",
				"from":      "0x1234567890123456789012345678901234567890",
			},
		},
		{
			name: "missing_token_in",
			args: map[string]interface{}{
				"chain":     "ethereum",
				"token_out": "0xC02aaA39b223FE65608aaC29B7b0Ed646Cf8F7D5",
				"amount_in": "1000.0",
				"from":      "0x1234567890123456789012345678901234567890",
			},
		},
		{
			name: "missing_token_out",
			args: map[string]interface{}{
				"chain":     "ethereum",
				"token_in":  "0xA0b86a33E6441e064c7d56Fb65dd0E8F8b764C4e",
				"amount_in": "1000.0",
				"from":      "0x1234567890123456789012345678901234567890",
			},
		},
		{
			name: "missing_amounts",
			args: map[string]interface{}{
				"chain":     "ethereum",
				"token_in":  "0xA0b86a33E6441e064c7d56Fb65dd0E8F8b764C4e",
				"token_out": "0xC02aaA39b223FE65608aaC29B7b0Ed646Cf8F7D5",
				"from":      "0x1234567890123456789012345678901234567890",
			},
		},
		{
			name: "both_amounts_specified",
			args: map[string]interface{}{
				"chain":      "ethereum",
				"token_in":   "0xA0b86a33E6441e064c7d56Fb65dd0E8F8b764C4e",
				"token_out":  "0xC02aaA39b223FE65608aaC29B7b0Ed646Cf8F7D5",
				"amount_in":  "1000.0",
				"amount_out": "500.0",
				"from":       "0x1234567890123456789012345678901234567890",
			},
		},
		{
			name: "unsupported_chain",
			args: map[string]interface{}{
				"chain":     "polygon",
				"token_in":  "0xA0b86a33E6441e064c7d56Fb65dd0E8F8b764C4e",
				"token_out": "0xC02aaA39b223FE65608aaC29B7b0Ed646Cf8F7D5",
				"amount_in": "1000.0",
				"from":      "0x1234567890123456789012345678901234567890",
			},
		},
		{
			name: "invalid_slippage",
			args: map[string]interface{}{
				"chain":              "ethereum",
				"token_in":           "0xA0b86a33E6441e064c7d56Fb65dd0E8F8b764C4e",
				"token_out":          "0xC02aaA39b223FE65608aaC29B7b0Ed646Cf8F7D5",
				"amount_in":          "1000.0",
				"from":               "0x1234567890123456789012345678901234567890",
				"slippage_tolerance": 100.0, // Too high
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := client.CallTool("swap_tokens", tc.args)
			require.NoError(t, err, "MCP call should not error")
			require.NotNil(t, result, "result should not be nil")
			// The tool should return an error message in the result, not throw an error
		})
	}
}

func TestSwapTokensToolDEXSelection(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	client := testEnv.GetMcpClient()
	require.NotNil(t, client, "MCP client should not be nil")

	require.NoError(t, client.Initialize(ctx), "failed to initialize MCP client")

	// Test with explicit Uniswap selection
	args := map[string]interface{}{
		"chain":     "ethereum",
		"token_in":  "0xA0b86a33E6441e064c7d56Fb65dd0E8F8b764C4e", // USDC
		"token_out": "0xC02aaA39b223FE65608aaC29B7b0Ed646Cf8F7D5", // WETH
		"amount_in": "1000.0",
		"from":      "0x1234567890123456789012345678901234567890",
		"dex":       "uniswap_v2",
	}
	result, err := client.CallTool("swap_tokens", args)
	require.NoError(t, err, "failed to call swap_tokens tool with explicit DEX")
	require.NotNil(t, result, "swap_tokens tool result should not be nil")

	// Test with unsupported DEX on wrong chain
	bscArgs := map[string]interface{}{
		"chain":     "bsc",
		"token_in":  "0xA0b86a33E6441e064c7d56Fb65dd0E8F8b764C4e",
		"token_out": "0xC02aaA39b223FE65608aaC29B7b0Ed646Cf8F7D5",
		"amount_in": "1000.0",
		"from":      "0x1234567890123456789012345678901234567890",
		"dex":       "uniswap", // Uniswap doesn't support BSC
	}
	bscResult, err := client.CallTool("swap_tokens", bscArgs)
	require.NoError(t, err, "MCP call should not error")
	require.NotNil(t, bscResult, "result should not be nil")
	// Should return error message about unsupported chain for this DEX
}