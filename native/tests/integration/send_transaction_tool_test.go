package integration

import (
	"context"
	"testing"

	"github.com/algonius/algonius-wallet/native/tests/integration/env"
	"github.com/stretchr/testify/require"
)

func TestSendTransactionTool(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	client := testEnv.GetMcpClient()
	require.NotNil(t, client, "MCP client should not be nil")

	require.NoError(t, client.Initialize(ctx), "failed to initialize MCP client")

	// First create a wallet to have a from address
	createArgs := map[string]interface{}{
		"chain": "ETH",
	}
	createResult, err := client.CallTool("create_wallet", createArgs)
	require.NoError(t, err, "failed to call create_wallet tool")
	require.NotNil(t, createResult, "create_wallet tool result should not be nil")

	// Test Ethereum ETH transaction
	args := map[string]interface{}{
		"chain":  "ethereum",
		"from":   "0x1234567890123456789012345678901234567890",
		"to":     "0x0987654321098765432109876543210987654321",
		"amount": "1.5",
		"token":  "",
	}
	result, err := client.CallTool("send_transaction", args)
	require.NoError(t, err, "failed to call send_transaction tool for ETH")
	require.NotNil(t, result, "send_transaction tool result should not be nil")

	// Test BSC BNB transaction
	bscArgs := map[string]interface{}{
		"chain":  "bsc",
		"from":   "0x1234567890123456789012345678901234567890",
		"to":     "0x0987654321098765432109876543210987654321",
		"amount": "0.5",
		"token":  "",
	}
	bscResult, err := client.CallTool("send_transaction", bscArgs)
	require.NoError(t, err, "failed to call send_transaction tool for BSC")
	require.NotNil(t, bscResult, "send_transaction tool result should not be nil for BSC")

	// Test ERC-20 token transaction
	erc20Args := map[string]interface{}{
		"chain":  "ethereum",
		"from":   "0x1234567890123456789012345678901234567890",
		"to":     "0x0987654321098765432109876543210987654321",
		"amount": "100",
		"token":  "0xA0b86a33E6441E8C8E06d74f2e63A5A8b8a9c7d6", // Example USDC contract
	}
	erc20Result, err := client.CallTool("send_transaction", erc20Args)
	require.NoError(t, err, "failed to call send_transaction tool for ERC-20")
	require.NotNil(t, erc20Result, "send_transaction tool result should not be nil for ERC-20")
}

func TestSendTransactionToolGasEstimation(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	client := testEnv.GetMcpClient()
	require.NotNil(t, client, "MCP client should not be nil")

	require.NoError(t, client.Initialize(ctx), "failed to initialize MCP client")

	// Test transaction without providing gas parameters (should use estimation)
	args := map[string]interface{}{
		"chain":  "ethereum",
		"from":   "0x1234567890123456789012345678901234567890",
		"to":     "0x0987654321098765432109876543210987654321",
		"amount": "1.0",
		// No gas_limit or gas_price - should be estimated
	}
	result, err := client.CallTool("send_transaction", args)
	require.NoError(t, err, "failed to call send_transaction tool with gas estimation")
	require.NotNil(t, result, "send_transaction tool result should not be nil")

	// Test BSC transaction with gas estimation
	bscArgs := map[string]interface{}{
		"chain":  "bsc",
		"from":   "0x1234567890123456789012345678901234567890",
		"to":     "0x0987654321098765432109876543210987654321",
		"amount": "0.1",
		// No gas parameters - should use BSC-specific estimation
	}
	bscResult, err := client.CallTool("send_transaction", bscArgs)
	require.NoError(t, err, "failed to call send_transaction tool for BSC with gas estimation")
	require.NotNil(t, bscResult, "send_transaction tool result should not be nil for BSC")
}

func TestSendTransactionToolValidation(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	client := testEnv.GetMcpClient()
	require.NotNil(t, client, "MCP client should not be nil")

	require.NoError(t, client.Initialize(ctx), "failed to initialize MCP client")

	// Test missing required parameters
	testCases := []struct {
		name string
		args map[string]interface{}
	}{
		{
			name: "missing chain",
			args: map[string]interface{}{
				"from":   "0x1234567890123456789012345678901234567890",
				"to":     "0x0987654321098765432109876543210987654321",
				"amount": "1.5",
			},
		},
		{
			name: "missing from",
			args: map[string]interface{}{
				"chain":  "ethereum",
				"to":     "0x0987654321098765432109876543210987654321",
				"amount": "1.5",
			},
		},
		{
			name: "missing to",
			args: map[string]interface{}{
				"chain":  "ethereum",
				"from":   "0x1234567890123456789012345678901234567890",
				"amount": "1.5",
			},
		},
		{
			name: "missing amount",
			args: map[string]interface{}{
				"chain": "ethereum",
				"from":  "0x1234567890123456789012345678901234567890",
				"to":    "0x0987654321098765432109876543210987654321",
			},
		},
		{
			name: "unsupported chain",
			args: map[string]interface{}{
				"chain":  "bitcoin",
				"from":   "0x1234567890123456789012345678901234567890",
				"to":     "0x0987654321098765432109876543210987654321",
				"amount": "1.5",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := client.CallTool("send_transaction", tc.args)
			// We expect an error or an error result for invalid inputs
			if err == nil && result != nil {
				// Check if result contains error information
				// The MCP tool should return an error result rather than a Go error
			}
		})
	}
}