// SPDX-License-Identifier: Apache-2.0
package integration

import (
	"context"
	"testing"

	"github.com/algonius/algonius-wallet/native/tests/integration/env"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImportWallet_NativeMessaging_Integration(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	nativeMsg := testEnv.GetNativeMsg()
	require.NotNil(t, nativeMsg, "Native messaging should not be nil")

	// Test successful wallet import
	t.Run("successful import via native messaging", func(t *testing.T) {
		params := map[string]interface{}{
			"mnemonic": "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
			"password": "test_password_123",
			"chain":    "ethereum",
		}

		response, err := nativeMsg.RpcRequest(ctx, "import_wallet", params)
		require.NoError(t, err, "import_wallet RPC should succeed")
		require.NotNil(t, response, "Response should not be nil")

		// Extract result data from response
		result, exists := response["result"]
		require.True(t, exists, "Response should contain result")
		resultData, ok := result.(map[string]interface{})
		require.True(t, ok, "Result should be a map")

		// Verify address
		address, exists := resultData["address"]
		require.True(t, exists, "Result should contain address")
		addressStr, ok := address.(string)
		require.True(t, ok, "Address should be a string")
		assert.True(t, len(addressStr) == 42, "Ethereum address should be 42 characters")
		assert.True(t, addressStr[:2] == "0x", "Ethereum address should start with 0x")

		// Verify public key
		publicKey, exists := resultData["publicKey"]
		require.True(t, exists, "Result should contain publicKey")
		publicKeyStr, ok := publicKey.(string)
		require.True(t, ok, "Public key should be a string")
		assert.True(t, len(publicKeyStr) > 0, "Public key should not be empty")
		assert.True(t, publicKeyStr[:2] == "0x", "Public key should start with 0x")

		// Verify importedAt timestamp
		importedAt, exists := resultData["importedAt"]
		require.True(t, exists, "Result should contain importedAt")
		importedAtFloat, ok := importedAt.(float64)
		require.True(t, ok, "ImportedAt should be a number")
		assert.Greater(t, int64(importedAtFloat), int64(0), "ImportedAt timestamp should be positive")
	})

	// Test BSC chain support
	t.Run("BSC chain support via native messaging", func(t *testing.T) {
		params := map[string]interface{}{
			"mnemonic": "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon cactus",
			"password": "test_password_bsc",
			"chain":    "bsc",
		}

		response, err := nativeMsg.RpcRequest(ctx, "import_wallet", params)
		require.NoError(t, err, "import_wallet RPC with BSC should succeed")
		require.NotNil(t, response, "Response should not be nil")

		// Extract result data from response
		result, exists := response["result"]
		require.True(t, exists, "Response should contain result")
		resultData, ok := result.(map[string]interface{})
		require.True(t, ok, "Result should be a map")

		// Verify result exists and has required fields
		address, exists := resultData["address"]
		require.True(t, exists, "Response should contain address")
		require.NotNil(t, address, "Address should not be nil")
	})

	// Test Solana chain support
	t.Run("Solana chain support via native messaging", func(t *testing.T) {
		params := map[string]interface{}{
			"mnemonic": "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art",
			"password": "test_password_solana",
			"chain":    "solana",
		}

		response, err := nativeMsg.RpcRequest(ctx, "import_wallet", params)
		require.NoError(t, err, "import_wallet RPC with Solana should succeed")
		require.NotNil(t, response, "Response should not be nil")

		// Extract result data from response
		result, exists := response["result"]
		require.True(t, exists, "Response should contain result")
		resultData, ok := result.(map[string]interface{})
		require.True(t, ok, "Result should be a map")

		// Verify result exists and has required fields
		address, exists := resultData["address"]
		require.True(t, exists, "Response should contain address")
		require.NotNil(t, address, "Address should not be nil")
		
		// Verify Solana address format (should be Base58 encoded, not start with 0x)
		addressStr, ok := address.(string)
		require.True(t, ok, "Address should be a string")
		assert.True(t, len(addressStr) > 32, "Solana address should be > 32 characters")
		assert.False(t, addressStr[:2] == "0x", "Solana address should not start with 0x")
	})

	// Test validation errors
	t.Run("validation errors via native messaging", func(t *testing.T) {
		tests := []struct {
			name           string
			params         map[string]interface{}
			expectedSubstr string
		}{
			{
				name: "invalid mnemonic",
				params: map[string]interface{}{
					"mnemonic": "invalid mnemonic phrase",
					"password": "password123",
					"chain":    "ethereum",
				},
				expectedSubstr: "invalid mnemonic",
			},
			{
				name: "weak password",
				params: map[string]interface{}{
					"mnemonic": "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
					"password": "123",
					"chain":    "ethereum",
				},
				expectedSubstr: "weak password",
			},
			{
				name: "unsupported chain",
				params: map[string]interface{}{
					"mnemonic": "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
					"password": "password123",
					"chain":    "bitcoin",
				},
				expectedSubstr: "unsupported chain",
			},
			{
				name: "empty mnemonic",
				params: map[string]interface{}{
					"mnemonic": "",
					"password": "password123",
					"chain":    "ethereum",
				},
				expectedSubstr: "required",
			},
			{
				name: "empty password",
				params: map[string]interface{}{
					"mnemonic": "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
					"password": "",
					"chain":    "ethereum",
				},
				expectedSubstr: "required",
			},
			{
				name: "empty chain",
				params: map[string]interface{}{
					"mnemonic": "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
					"password": "password123",
					"chain":    "",
				},
				expectedSubstr: "required",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				response, err := nativeMsg.RpcRequest(ctx, "import_wallet", tt.params)
				require.NoError(t, err, "RPC transport should succeed")
				require.NotNil(t, response, "Response should not be nil")
				
				// Should have error in response structure
				errorInfo, exists := response["error"]
				require.True(t, exists, "Response should contain error")
				require.NotNil(t, errorInfo, "Error should not be nil")
				
				errorMap, ok := errorInfo.(map[string]interface{})
				require.True(t, ok, "Error should be a map")
				
				// Check error message contains expected substring
				message, exists := errorMap["message"]
				require.True(t, exists, "Error should have message")
				messageStr, ok := message.(string)
				require.True(t, ok, "Error message should be a string")
				assert.Contains(t, messageStr, tt.expectedSubstr, "Error message should contain expected substring")
			})
		}
	})

	// Test that import_wallet is not available via MCP tools (security requirement)
	t.Run("import_wallet not available via MCP", func(t *testing.T) {
		mcpClient := testEnv.GetMcpClient()
		require.NotNil(t, mcpClient, "MCP client should not be nil")

		// Try to call import_wallet as a tool (should fail)
		args := map[string]interface{}{
			"mnemonic": "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
			"password": "test_password_123",
			"chain":    "ethereum",
		}

		_, err := mcpClient.CallTool("import_wallet", args)
		require.Error(t, err, "import_wallet should not be available as MCP tool")
		assert.Contains(t, err.Error(), "tool not found", "Error should indicate tool not found")
	})
}