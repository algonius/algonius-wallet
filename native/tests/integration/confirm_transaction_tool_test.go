package integration

import (
	"context"
	"strings"
	"testing"

	"github.com/algonius/algonius-wallet/native/tests/integration/env"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

func TestConfirmTransactionTool(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	client := testEnv.GetMcpClient()
	require.NotNil(t, client, "MCP client should not be nil")

	require.NoError(t, client.Initialize(ctx), "failed to initialize MCP client")

	// Test successful ethereum transaction confirmation
	args := map[string]interface{}{
		"chain":   "ethereum",
		"tx_hash": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
	}
	result, err := client.CallTool("confirm_transaction", args)
	require.NoError(t, err, "failed to call confirm_transaction tool")
	require.NotNil(t, result, "confirm_transaction tool result should not be nil")
	require.False(t, result.IsError, "result should not be an error")

	// Extract text content
	textContent, ok := mcp.AsTextContent(result.Content[0])
	require.True(t, ok, "content should be TextContent type")
	responseText := textContent.Text

	// Verify the response contains expected markdown format
	require.Contains(t, responseText, "### Transaction Confirmation Status")
	require.Contains(t, responseText, "**Transaction Hash**:")
	require.Contains(t, responseText, "**Chain**: `ETH`")
	require.Contains(t, responseText, "**Status**:")
	require.Contains(t, responseText, "**Confirmations**:")
}

func TestConfirmTransactionToolBSC(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	client := testEnv.GetMcpClient()
	require.NotNil(t, client, "MCP client should not be nil")

	require.NoError(t, client.Initialize(ctx), "failed to initialize MCP client")

	// Test BSC transaction confirmation
	args := map[string]interface{}{
		"chain":   "bsc",
		"tx_hash": "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
	}
	result, err := client.CallTool("confirm_transaction", args)
	require.NoError(t, err, "failed to call confirm_transaction tool for BSC")
	require.NotNil(t, result, "confirm_transaction tool result should not be nil")
	require.False(t, result.IsError, "result should not be an error")

	// Extract text content
	textContent, ok := mcp.AsTextContent(result.Content[0])
	require.True(t, ok, "content should be TextContent type")
	responseText := textContent.Text

	// Verify BSC-specific response
	require.Contains(t, responseText, "**Chain**: `BSC`")
	require.Contains(t, responseText, "### Transaction Confirmation Status")
}

func TestConfirmTransactionToolWithCustomConfirmations(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	client := testEnv.GetMcpClient()
	require.NotNil(t, client, "MCP client should not be nil")

	require.NoError(t, client.Initialize(ctx), "failed to initialize MCP client")

	// Test with custom required confirmations
	args := map[string]interface{}{
		"chain":                  "ethereum",
		"tx_hash":                "0x1111111111111111111111111111111111111111111111111111111111111111",
		"required_confirmations": 12,
	}
	result, err := client.CallTool("confirm_transaction", args)
	require.NoError(t, err, "failed to call confirm_transaction tool with custom confirmations")
	require.NotNil(t, result, "confirm_transaction tool result should not be nil")
	require.False(t, result.IsError, "result should not be an error")

	// Extract text content
	textContent, ok := mcp.AsTextContent(result.Content[0])
	require.True(t, ok, "content should be TextContent type")
	responseText := textContent.Text

	// Verify custom confirmation threshold is reflected in response
	require.Contains(t, responseText, "**Required Confirmations**: `12`")
}

func TestConfirmTransactionToolHashNormalization(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	client := testEnv.GetMcpClient()
	require.NotNil(t, client, "MCP client should not be nil")

	require.NoError(t, client.Initialize(ctx), "failed to initialize MCP client")

	// Test hash without 0x prefix (should be normalized)
	args := map[string]interface{}{
		"chain":   "ethereum",
		"tx_hash": "2222222222222222222222222222222222222222222222222222222222222222",
	}
	result, err := client.CallTool("confirm_transaction", args)
	require.NoError(t, err, "failed to call confirm_transaction tool with hash without 0x prefix")
	require.NotNil(t, result, "confirm_transaction tool result should not be nil")
	require.False(t, result.IsError, "result should not be an error")

	// Extract text content
	textContent, ok := mcp.AsTextContent(result.Content[0])
	require.True(t, ok, "content should be TextContent type")
	responseText := textContent.Text

	// Verify the hash is normalized to include 0x prefix
	require.Contains(t, responseText, "0x2222222222222222222222222222222222222222222222222222222222222222")
}

func TestConfirmTransactionToolDifferentTransactionStates(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	client := testEnv.GetMcpClient()
	require.NotNil(t, client, "MCP client should not be nil")

	require.NoError(t, client.Initialize(ctx), "failed to initialize MCP client")

	// Test different transaction states based on mock implementation logic
	testCases := []struct {
		name   string
		txHash string
		// Mock implementation uses last byte of hash to determine state:
		// lastByte%10 == 0 -> failed (10% chance)
		// lastByte%3 == 0 -> pending (~33% chance)  
		// default -> confirmed (~57% chance)
	}{
		{
			name:   "failed_transaction",
			txHash: "0x1111111111111111111111111111111111111111111111111111111111111110", // ends in 0 -> failed
		},
		{
			name:   "pending_transaction", 
			txHash: "0x1111111111111111111111111111111111111111111111111111111111111113", // ends in 3 -> pending
		},
		{
			name:   "confirmed_transaction",
			txHash: "0x1111111111111111111111111111111111111111111111111111111111111111", // ends in 1 -> confirmed
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			args := map[string]interface{}{
				"chain":   "ethereum",
				"tx_hash": tc.txHash,
			}
			result, err := client.CallTool("confirm_transaction", args)
			require.NoError(t, err, "failed to call confirm_transaction tool for %s", tc.name)
			require.NotNil(t, result, "confirm_transaction tool result should not be nil")
			require.False(t, result.IsError, "result should not be an error")

			// Extract text content
			textContent, ok := mcp.AsTextContent(result.Content[0])
			require.True(t, ok, "content should be TextContent type")
			responseText := textContent.Text

			// Verify response contains transaction status
			require.Contains(t, responseText, "**Status**:")
			// All should have valid markdown format
			require.Contains(t, responseText, "### Transaction Confirmation Status")
		})
	}
}

func TestConfirmTransactionToolValidation(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	client := testEnv.GetMcpClient()
	require.NotNil(t, client, "MCP client should not be nil")

	require.NoError(t, client.Initialize(ctx), "failed to initialize MCP client")

	testCases := []struct {
		name        string
		args        map[string]interface{}
		expectError bool
		errorText   string
	}{
		{
			name: "missing_chain",
			args: map[string]interface{}{
				"tx_hash": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			},
			expectError: true,
			errorText:   "missing or invalid 'chain' parameter",
		},
		{
			name: "missing_tx_hash",
			args: map[string]interface{}{
				"chain": "ethereum",
			},
			expectError: true,
			errorText:   "missing or invalid 'tx_hash' parameter",
		},
		{
			name: "unsupported_chain",
			args: map[string]interface{}{
				"chain":   "bitcoin",
				"tx_hash": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			},
			expectError: true,
			errorText:   "unsupported chain",
		},
		{
			name: "invalid_hash_length",
			args: map[string]interface{}{
				"chain":   "ethereum",
				"tx_hash": "0x123", // Too short
			},
			expectError: true,
			errorText:   "Failed to check transaction confirmation",
		},
		{
			name: "invalid_hash_format",
			args: map[string]interface{}{
				"chain":   "ethereum", 
				"tx_hash": "0xGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGG", // Invalid hex
			},
			expectError: true,
			errorText:   "Failed to check transaction confirmation",
		},
	}

		for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := client.CallTool("confirm_transaction", tc.args)
			
			if tc.expectError {
				// Should not return a Go error, but should return an MCP error result
				require.NoError(t, err, "should not return Go error")
				require.NotNil(t, result, "result should not be nil")
				require.True(t, result.IsError, "result should be an error")
				
				// Extract text content for error messages
				textContent, ok := mcp.AsTextContent(result.Content[0])
				require.True(t, ok, "content should be TextContent type")
				responseText := textContent.Text
				
				require.Contains(t, strings.ToLower(responseText), strings.ToLower(tc.errorText))
			} else {
				require.NoError(t, err, "should not return error for valid input")
				require.NotNil(t, result, "result should not be nil")
				require.False(t, result.IsError, "result should not be an error")
			}
		})
	}
}

func TestConfirmTransactionToolChainNameVariants(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	client := testEnv.GetMcpClient()
	require.NotNil(t, client, "MCP client should not be nil")

	require.NoError(t, client.Initialize(ctx), "failed to initialize MCP client")

	// Test different chain name formats
	testCases := []struct {
		chainName     string
		expectedChain string
	}{
		{"ethereum", "ETH"},
		{"eth", "ETH"},
		{"ETH", "ETH"},
		{"Ethereum", "ETH"},
		{"bsc", "BSC"},
		{"BSC", "BSC"},
		{"binance", "BSC"},
		{"binance-smart-chain", "BSC"},
	}

	txHash := "0x4444444444444444444444444444444444444444444444444444444444444444"

	for _, tc := range testCases {
		t.Run("chain_name_"+tc.chainName, func(t *testing.T) {
			args := map[string]interface{}{
				"chain":   tc.chainName,
				"tx_hash": txHash,
			}
			result, err := client.CallTool("confirm_transaction", args)
			require.NoError(t, err, "failed to call confirm_transaction tool for chain %s", tc.chainName)
			require.NotNil(t, result, "confirm_transaction tool result should not be nil")
			require.False(t, result.IsError, "result should not be an error for valid chain name")

			// Extract text content
			textContent, ok := mcp.AsTextContent(result.Content[0])
			require.True(t, ok, "content should be TextContent type")
			responseText := textContent.Text

			// Verify the normalized chain name appears in response
			require.Contains(t, responseText, "**Chain**: `"+tc.expectedChain+"`")
		})
	}
}