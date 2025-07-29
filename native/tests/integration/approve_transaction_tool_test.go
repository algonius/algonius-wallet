package integration

import (
	"context"
	"strings"
	"testing"

	"github.com/algonius/algonius-wallet/native/tests/integration/env"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApproveTransactionTool(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	client := testEnv.GetMcpClient()
	require.NotNil(t, client, "MCP client should not be nil")

	require.NoError(t, client.Initialize(ctx), "failed to initialize MCP client")

	// First, create a wallet to have a proper context
	createArgs := map[string]interface{}{
		"chain": "ETH",
	}
	createResult, err := client.CallTool("create_wallet", createArgs)
	require.NoError(t, err, "failed to create wallet")
	require.NotNil(t, createResult, "create wallet result should not be nil")

	// Test approve transaction with valid transaction hash (simulation)
	args := map[string]interface{}{
		"transaction_hash": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		"action":           "approve",
	}
	result, err := client.CallTool("approve_transaction", args)
	require.NoError(t, err, "failed to call approve_transaction tool")
	require.NotNil(t, result, "approve_transaction tool result should not be nil")

	// Since we don't have actual pending transactions in test mode,
	// this should return an error about transaction not found
	if result.IsError {
		textContent := getTextContent(result)
		assert.Contains(t, textContent, "not found", "should indicate transaction not found")
	}
}

func TestApproveTransactionToolReject(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	client := testEnv.GetMcpClient()
	require.NotNil(t, client, "MCP client should not be nil")

	require.NoError(t, client.Initialize(ctx), "failed to initialize MCP client")

	// First, create a wallet
	createArgs := map[string]interface{}{
		"chain": "ETH",
	}
	createResult, err := client.CallTool("create_wallet", createArgs)
	require.NoError(t, err, "failed to create wallet")
	require.NotNil(t, createResult, "create wallet result should not be nil")

	// Test reject transaction with valid parameters
	args := map[string]interface{}{
		"transaction_hash": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		"action":           "reject",
		"reason":           "Transaction appears suspicious",
	}
	result, err := client.CallTool("approve_transaction", args)
	require.NoError(t, err, "failed to call approve_transaction tool for rejection")
	require.NotNil(t, result, "approve_transaction tool result should not be nil")

	// Should return error about transaction not found (expected in test environment)
	if result.IsError {
		textContent := getTextContent(result)
		assert.Contains(t, textContent, "not found", "should indicate transaction not found")
	}
}

func TestApproveTransactionToolParameterValidation(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	client := testEnv.GetMcpClient()
	require.NotNil(t, client, "MCP client should not be nil")

	require.NoError(t, client.Initialize(ctx), "failed to initialize MCP client")

	testCases := []struct {
		name           string
		args           map[string]interface{}
		expectedError  string
		description    string
	}{
		{
			name:          "missing_transaction_hash",
			args:          map[string]interface{}{"action": "approve"},
			expectedError: "transaction_hash",
			description:   "should require transaction_hash parameter",
		},
		{
			name:          "missing_action",
			args:          map[string]interface{}{"transaction_hash": "0x123"},
			expectedError: "action",
			description:   "should require action parameter",
		},
		{
			name: "invalid_action",
			args: map[string]interface{}{
				"transaction_hash": "0x123",
				"action":           "invalid",
			},
			expectedError: "action",
			description:   "should validate action parameter",
		},
		{
			name: "reject_without_reason",
			args: map[string]interface{}{
				"transaction_hash": "0x123",
				"action":           "reject",
			},
			expectedError: "reason",
			description:   "should require reason when rejecting",
		},
		{
			name: "empty_transaction_hash",
			args: map[string]interface{}{
				"transaction_hash": "",
				"action":           "approve",
			},
			expectedError: "not found",
			description:   "empty transaction hash should result in transaction not found",
		},
		{
			name: "empty_action",
			args: map[string]interface{}{
				"transaction_hash": "0x123",
				"action":           "",
			},
			expectedError: "action",
			description:   "should not accept empty action",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := client.CallTool("approve_transaction", tc.args)
			require.NoError(t, err, "tool call should not return Go error")
			require.NotNil(t, result, "result should not be nil")
			require.True(t, result.IsError, tc.description)

			textContent := getTextContent(result)
			assert.Contains(t, strings.ToLower(textContent), strings.ToLower(tc.expectedError), 
				"error message should contain expected field name")
		})
	}
}

func TestApproveTransactionToolActionTypes(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	client := testEnv.GetMcpClient()
	require.NotNil(t, client, "MCP client should not be nil")

	require.NoError(t, client.Initialize(ctx), "failed to initialize MCP client")

	testCases := []struct {
		name   string
		action string
		reason string
		valid  bool
	}{
		{
			name:   "approve_action",
			action: "approve",
			reason: "", // reason not required for approve
			valid:  true,
		},
		{
			name:   "reject_action_with_reason",
			action: "reject",
			reason: "Suspicious transaction",
			valid:  true,
		},
		{
			name:   "APPROVE_uppercase",
			action: "APPROVE",
			reason: "",
			valid:  false, // should be case sensitive
		},
		{
			name:   "reject_uppercase",
			action: "REJECT",
			reason: "Test reason",
			valid:  false, // should be case sensitive
		},
		{
			name:   "mixed_case_approve",
			action: "Approve",
			reason: "",
			valid:  false, // should be case sensitive
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			args := map[string]interface{}{
				"transaction_hash": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
				"action":           tc.action,
			}
			if tc.reason != "" {
				args["reason"] = tc.reason
			}

			result, err := client.CallTool("approve_transaction", args)
			require.NoError(t, err, "tool call should not return Go error")
			require.NotNil(t, result, "result should not be nil")

			if tc.valid {
				// Even valid requests will fail in test environment due to no pending transactions
				// But they should fail with "not found" rather than parameter validation error
				if result.IsError {
					textContent := getTextContent(result)
					assert.Contains(t, textContent, "not found", 
						"valid parameters should fail on transaction not found, not parameter validation")
				}
			} else {
				require.True(t, result.IsError, "invalid action should return error")
				textContent := getTextContent(result)
				assert.Contains(t, strings.ToLower(textContent), "action", 
					"error should be about action parameter")
			}
		})
	}
}

func TestApproveTransactionToolTransactionHashFormats(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	client := testEnv.GetMcpClient()
	require.NotNil(t, client, "MCP client should not be nil")

	require.NoError(t, client.Initialize(ctx), "failed to initialize MCP client")

	testCases := []struct {
		name            string
		transactionHash string
		description     string
	}{
		{
			name:            "standard_eth_hash",
			transactionHash: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			description:     "standard Ethereum transaction hash format",
		},
		{
			name:            "hash_without_0x_prefix",
			transactionHash: "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			description:     "transaction hash without 0x prefix",
		},
		{
			name:            "uppercase_hash",
			transactionHash: "0x1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF",
			description:     "uppercase transaction hash",
		},
		{
			name:            "mixed_case_hash",
			transactionHash: "0x1234567890AbCdEf1234567890aBcDeF1234567890AbCdEf1234567890aBcDeF",
			description:     "mixed case transaction hash",
		},
		{
			name:            "short_hash",
			transactionHash: "0x123456",
			description:     "shorter transaction hash",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			args := map[string]interface{}{
				"transaction_hash": tc.transactionHash,
				"action":           "approve",
			}

			result, err := client.CallTool("approve_transaction", args)
			require.NoError(t, err, "tool call should not return Go error")
			require.NotNil(t, result, "result should not be nil")

			// All these should pass parameter validation but fail on transaction not found
			if result.IsError {
				textContent := getTextContent(result)
				assert.Contains(t, textContent, "not found", 
					"should fail on transaction not found, not parameter format validation")
			}
		})
	}
}

func TestApproveTransactionToolReasonValidation(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	client := testEnv.GetMcpClient()
	require.NotNil(t, client, "MCP client should not be nil")

	require.NoError(t, client.Initialize(ctx), "failed to initialize MCP client")

	testCases := []struct {
		name          string
		reason        string
		expectError   bool
		description   string
	}{
		{
			name:        "simple_reason",
			reason:      "Transaction looks suspicious",
			expectError: false,
			description: "simple rejection reason should be accepted",
		},
		{
			name:        "detailed_reason",
			reason:      "Transaction amount is unusually high and destination address is not in whitelist",
			expectError: false,
			description: "detailed rejection reason should be accepted",
		},
		{
			name:        "reason_with_special_chars",
			reason:      "Reason: Transaction contains special chars & symbols! @#$%",
			expectError: false,
			description: "reason with special characters should be accepted",
		},
		{
			name:        "empty_reason",
			reason:      "",
			expectError: true,
			description: "empty reason should be rejected for reject action",
		},
		{
			name:        "whitespace_only_reason",
			reason:      "   ",
			expectError: false,
			description: "whitespace-only reason should be accepted as valid string",
		},
		{
			name:        "unicode_reason",
			reason:      "交易看起来可疑 - Transaction looks suspicious",
			expectError: false,
			description: "unicode characters in reason should be accepted",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			args := map[string]interface{}{
				"transaction_hash": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
				"action":           "reject",
				"reason":           tc.reason,
			}

			result, err := client.CallTool("approve_transaction", args)
			require.NoError(t, err, "tool call should not return Go error")
			require.NotNil(t, result, "result should not be nil")

			if tc.expectError {
				require.True(t, result.IsError, tc.description)
				textContent := getTextContent(result)
				assert.Contains(t, strings.ToLower(textContent), "reason", 
					"error should mention reason field")
			} else {
				// Valid reasons should pass validation but fail on transaction not found
				if result.IsError {
					textContent := getTextContent(result)
					assert.Contains(t, textContent, "not found", 
						"should fail on transaction not found, not reason validation")
				}
			}
		})
	}
}

func TestApproveTransactionToolSecurityValidation(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	client := testEnv.GetMcpClient()
	require.NotNil(t, client, "MCP client should not be nil")

	require.NoError(t, client.Initialize(ctx), "failed to initialize MCP client")

	securityTestCases := []struct {
		name            string
		transactionHash string
		action          string
		reason          string
		description     string
	}{
		{
			name:            "sql_injection_attempt",
			transactionHash: "0x123'; DROP TABLE transactions; --",
			action:          "approve",
			reason:          "",
			description:     "SQL injection attempt should not cause system error",
		},
		{
			name:            "script_injection_in_reason",
			transactionHash: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			action:          "reject",
			reason:          "<script>alert('xss')</script>",
			description:     "script injection in reason should be handled safely",
		},
		{
			name:            "command_injection_attempt",
			transactionHash: "0x123; rm -rf /",
			action:          "approve",
			reason:          "",
			description:     "command injection attempt should not cause system error",
		},
		{
			name:            "null_byte_injection",
			transactionHash: "0x123\x00malicious",
			action:          "approve",
			reason:          "",
			description:     "null byte injection should be handled safely",
		},
		{
			name:            "long_input_hash",
			transactionHash: strings.Repeat("a", 10000),
			action:          "approve",
			reason:          "",
			description:     "extremely long transaction hash should not cause system error",
		},
	}

	for _, tc := range securityTestCases {
		t.Run(tc.name, func(t *testing.T) {
			args := map[string]interface{}{
				"transaction_hash": tc.transactionHash,
				"action":           tc.action,
			}
			if tc.reason != "" {
				args["reason"] = tc.reason
			}

			result, err := client.CallTool("approve_transaction", args)
			require.NoError(t, err, tc.description)
			require.NotNil(t, result, "result should not be nil")

			// Security tests should not crash the system
			// They may return errors, but should be handled gracefully
			if result.IsError {
				textContent := getTextContent(result)
				// Should not contain system error traces or sensitive information
				assert.NotContains(t, textContent, "panic", "should not contain panic traces")
				assert.NotContains(t, textContent, "goroutine", "should not contain goroutine traces")
				assert.NotContains(t, textContent, "/Users/", "should not contain file paths")
			}
		})
	}
}

func TestApproveTransactionToolErrorHandling(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	client := testEnv.GetMcpClient()
	require.NotNil(t, client, "MCP client should not be nil")

	require.NoError(t, client.Initialize(ctx), "failed to initialize MCP client")

	errorTestCases := []struct {
		name           string
		args           map[string]interface{}
		expectedError  bool
		description    string
	}{
		{
			name: "malformed_json_like_hash",
			args: map[string]interface{}{
				"transaction_hash": `{"malformed": "json"}`,
				"action":           "approve",
			},
			expectedError: false, // should handle as normal string
			description:   "JSON-like strings should be treated as normal strings",
		},
		{
			name: "array_like_hash",
			args: map[string]interface{}{
				"transaction_hash": `["array", "elements"]`,
				"action":           "approve",
			},
			expectedError: false, // should handle as normal string
			description:   "array-like strings should be treated as normal strings",
		},
		{
			name: "numeric_string_hash",
			args: map[string]interface{}{
				"transaction_hash": "123456789",
				"action":           "approve",
			},
			expectedError: false,
			description:   "numeric strings should be accepted as transaction hashes",
		},
	}

	for _, tc := range errorTestCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := client.CallTool("approve_transaction", tc.args)
			require.NoError(t, err, tc.description)
			require.NotNil(t, result, "result should not be nil")

			if tc.expectedError {
				require.True(t, result.IsError, tc.description)
			} else {
				// Even if parameters are valid, should fail on transaction not found
				if result.IsError {
					textContent := getTextContent(result)
					assert.Contains(t, textContent, "not found", 
						"should fail gracefully with transaction not found")
				}
			}
		})
	}
}

func TestApproveTransactionToolContentValidation(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	client := testEnv.GetMcpClient()
	require.NotNil(t, client, "MCP client should not be nil")

	require.NoError(t, client.Initialize(ctx), "failed to initialize MCP client")

	// Test that the tool returns properly formatted content
	args := map[string]interface{}{
		"transaction_hash": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		"action":           "approve",
	}

	result, err := client.CallTool("approve_transaction", args)
	require.NoError(t, err, "failed to call approve_transaction tool")
	require.NotNil(t, result, "result should not be nil")

	// Verify content structure
	require.NotEmpty(t, result.Content, "result should have content")
	
	textContent, ok := mcp.AsTextContent(result.Content[0])
	require.True(t, ok, "result content should be text")
	require.NotEmpty(t, textContent.Text, "text content should not be empty")

	// Content should be markdown formatted
	text := textContent.Text
	if !result.IsError {
		// If successful, should contain markdown formatting
		assert.True(t, strings.Contains(text, "###") || strings.Contains(text, "##") || strings.Contains(text, "#"), 
			"successful response should contain markdown headers")
		assert.Contains(t, text, "**", "should contain bold markdown formatting")
	}
}

func TestApproveTransactionToolPerformance(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	client := testEnv.GetMcpClient()
	require.NotNil(t, client, "MCP client should not be nil")

	require.NoError(t, client.Initialize(ctx), "failed to initialize MCP client")

	// Test rapid succession of requests
	args := map[string]interface{}{
		"transaction_hash": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		"action":           "approve",
	}

	const numRequests = 10
	for i := 0; i < numRequests; i++ {
		result, err := client.CallTool("approve_transaction", args)
		require.NoError(t, err, "request %d should not return error", i+1)
		require.NotNil(t, result, "request %d result should not be nil", i+1)
		
		// Each request should be handled independently
		require.NotEmpty(t, result.Content, "request %d should have content", i+1)
	}
}