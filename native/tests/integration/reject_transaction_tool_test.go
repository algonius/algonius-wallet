package integration

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/algonius/algonius-wallet/native/tests/integration/env"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRejectTransactionTool(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	client := testEnv.GetMcpClient()
	require.NotNil(t, client, "MCP client should not be nil")

	require.NoError(t, client.Initialize(ctx), "failed to initialize MCP client")

	// Test basic rejection with single transaction
	args := map[string]interface{}{
		"transaction_ids": "0x1234567890abcdef1234567890abcdef12345678901234567890abcdef123456",
		"reason":          "suspicious_activity",
	}
	result, err := client.CallTool("reject_transaction", args)
	require.NoError(t, err, "failed to call reject_transaction tool")
	require.NotNil(t, result, "reject_transaction tool result should not be nil")
	require.False(t, result.IsError, "basic rejection should not return error")

	// Test rejection with multiple transactions
	args = map[string]interface{}{
		"transaction_ids": "0x1234567890abcdef1234567890abcdef12345678901234567890abcdef123456,0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
		"reason":          "high_gas_fee",
		"details":         "Gas fee exceeds 0.01 ETH threshold",
		"notify_user":     true,
		"audit_log":       true,
	}
	result, err = client.CallTool("reject_transaction", args)
	require.NoError(t, err, "failed to call reject_transaction tool with multiple IDs")
	require.NotNil(t, result, "reject_transaction tool result should not be nil")
	require.False(t, result.IsError, "multiple rejection should not return error")

	// Test rejection with all valid reasons
	validReasons := []string{
		"suspicious_activity",
		"high_gas_fee",
		"user_request",
		"security_concern",
		"duplicate_transaction",
	}
	
	for _, reason := range validReasons {
		t.Run(fmt.Sprintf("reason_%s", reason), func(t *testing.T) {
			args := map[string]interface{}{
				"transaction_ids": "0x1234567890abcdef1234567890abcdef12345678901234567890abcdef123456",
				"reason":          reason,
			}
			result, err := client.CallTool("reject_transaction", args)
			require.NoError(t, err, "failed to call reject_transaction tool")
			require.NotNil(t, result, "reject_transaction tool result should not be nil")
			require.False(t, result.IsError, "rejection with valid reason should not return error")
		})
	}

	// Test with notification enabled
	args = map[string]interface{}{
		"transaction_ids": "0x1234567890abcdef1234567890abcdef12345678901234567890abcdef123456",
		"reason":          "user_request",
		"notify_user":     true,
	}
	result, err = client.CallTool("reject_transaction", args)
	require.NoError(t, err, "failed to call reject_transaction tool with notification")
	require.NotNil(t, result, "reject_transaction tool result should not be nil")
	require.False(t, result.IsError, "rejection with notification should not return error")

	// Test with audit logging disabled
	args = map[string]interface{}{
		"transaction_ids": "0x1234567890abcdef1234567890abcdef12345678901234567890abcdef123456",
		"reason":          "user_request",
		"audit_log":       false,
	}
	result, err = client.CallTool("reject_transaction", args)
	require.NoError(t, err, "failed to call reject_transaction tool without audit log")
	require.NotNil(t, result, "reject_transaction tool result should not be nil")
	require.False(t, result.IsError, "rejection without audit log should not return error")
}

func TestRejectTransactionToolValidation(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	client := testEnv.GetMcpClient()
	require.NotNil(t, client, "MCP client should not be nil")

	require.NoError(t, client.Initialize(ctx), "failed to initialize MCP client")

	// Test missing transaction_ids parameter
	args := map[string]interface{}{
		"reason": "suspicious_activity",
	}
	result, err := client.CallTool("reject_transaction", args)
	require.NoError(t, err, "call should not return system error")
	require.NotNil(t, result, "result should not be nil")
	require.True(t, result.IsError, "should return error for missing transaction_ids")

	// Test missing reason parameter
	args = map[string]interface{}{
		"transaction_ids": "0x1234567890abcdef1234567890abcdef12345678901234567890abcdef123456",
	}
	result, err = client.CallTool("reject_transaction", args)
	require.NoError(t, err, "call should not return system error")
	require.NotNil(t, result, "result should not be nil")
	require.True(t, result.IsError, "should return error for missing reason")

	// Test empty transaction_ids
	args = map[string]interface{}{
		"transaction_ids": "",
		"reason":          "suspicious_activity",
	}
	result, err = client.CallTool("reject_transaction", args)
	require.NoError(t, err, "call should not return system error")
	require.NotNil(t, result, "result should not be nil")
	require.True(t, result.IsError, "should return error for empty transaction_ids")

	// Test invalid reason
	args = map[string]interface{}{
		"transaction_ids": "0x1234567890abcdef1234567890abcdef12345678901234567890abcdef123456",
		"reason":          "invalid_reason",
	}
	result, err = client.CallTool("reject_transaction", args)
	require.NoError(t, err, "call should not return system error")
	require.NotNil(t, result, "result should not be nil")
	require.True(t, result.IsError, "should return error for invalid reason")

	// Test whitespace-only transaction IDs
	args = map[string]interface{}{
		"transaction_ids": "   ,  ,   ",
		"reason":          "suspicious_activity",
	}
	result, err = client.CallTool("reject_transaction", args)
	require.NoError(t, err, "call should not return system error")
	require.NotNil(t, result, "result should not be nil")
	require.True(t, result.IsError, "should return error for whitespace-only transaction IDs")
}

func TestRejectTransactionToolContentValidation(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	client := testEnv.GetMcpClient()
	require.NotNil(t, client, "MCP client should not be nil")

	require.NoError(t, client.Initialize(ctx), "failed to initialize MCP client")

	// Test response format with successful rejection
	args := map[string]interface{}{
		"transaction_ids": "0x1234567890abcdef1234567890abcdef12345678901234567890abcdef123456",
		"reason":          "suspicious_activity",
		"details":         "Unusual transaction pattern detected",
		"notify_user":     true,
		"audit_log":       true,
	}
	result, err := client.CallTool("reject_transaction", args)
	require.NoError(t, err, "failed to call reject_transaction tool")
	require.NotNil(t, result, "reject_transaction tool result should not be nil")
	require.False(t, result.IsError, "successful rejection should not return error")

	// Validate response structure
	require.NotNil(t, result.Content, "result should have content")
	require.Len(t, result.Content, 1, "result should have one content item")

	// Validate markdown content format
	markdown := getTextContent(result)
	require.NotEmpty(t, markdown, "should have text content")
	require.True(t, strings.Contains(markdown, "### Transaction Rejection Results"), "should contain main heading")
	require.True(t, strings.Contains(markdown, "**Summary**"), "should contain summary section")
	require.True(t, strings.Contains(markdown, "Successfully rejected"), "should contain success count")
	require.True(t, strings.Contains(markdown, "**Rejection Details**"), "should contain rejection details")
	require.True(t, strings.Contains(markdown, "suspicious_activity"), "should contain the reason")
	require.True(t, strings.Contains(markdown, "Unusual transaction pattern detected"), "should contain details")
	require.True(t, strings.Contains(markdown, "#### Individual Results"), "should contain individual results section")

	// Test with multiple transactions
	args = map[string]interface{}{
		"transaction_ids": "0x1234567890abcdef1234567890abcdef12345678901234567890abcdef123456,0xnonexistent",
		"reason":          "high_gas_fee",
	}
	result, err = client.CallTool("reject_transaction", args)
	require.NoError(t, err, "failed to call reject_transaction tool")
	require.NotNil(t, result, "reject_transaction tool result should not be nil")
	require.False(t, result.IsError, "mixed results should not return error")

	markdown = getTextContent(result)
	require.True(t, strings.Contains(markdown, "Failed to reject"), "should show failure count for mixed results")
	require.True(t, strings.Contains(markdown, "✅ SUCCESS") || strings.Contains(markdown, "❌ FAILED"), "should show individual statuses")
}

func TestRejectTransactionToolSecurityValidation(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	client := testEnv.GetMcpClient()
	require.NotNil(t, client, "MCP client should not be nil")

	require.NoError(t, client.Initialize(ctx), "failed to initialize MCP client")

	// Test injection attempts
	injectionAttempts := []map[string]interface{}{
		{
			"transaction_ids": "'; DROP TABLE transactions; --",
			"reason":          "suspicious_activity",
		},
		{
			"transaction_ids": "0x123",
			"reason":          "<script>alert('xss')</script>",
		},
		{
			"transaction_ids": "../../../etc/passwd",
			"reason":          "user_request",
		},
		{
			"transaction_ids": strings.Repeat("0x", 1000),
			"reason":          "suspicious_activity",
		},
	}

	for i, args := range injectionAttempts {
		t.Run(fmt.Sprintf("injection_%d", i), func(t *testing.T) {
			result, err := client.CallTool("reject_transaction", args)
			// Should not panic or cause system errors
			require.NoError(t, err, "injection attempt should not cause system error")
			require.NotNil(t, result, "result should not be nil")
			// May return error or process gracefully, but should not crash
		})
	}

	// Test parameter type validation
	typeValidationTests := []map[string]interface{}{
		{
			"transaction_ids": 123,
			"reason":          "suspicious_activity",
		},
		{
			"transaction_ids": []string{"tx1", "tx2"},
			"reason":          "suspicious_activity",
		},
		{
			"transaction_ids": "0x123",
			"reason":          456,
		},
		{
			"transaction_ids": "0x123",
			"reason":          "suspicious_activity",
			"notify_user":     "not_a_boolean",
		},
	}

	for i, args := range typeValidationTests {
		t.Run(fmt.Sprintf("type_validation_%d", i), func(t *testing.T) {
			result, err := client.CallTool("reject_transaction", args)
			// Should handle type mismatches gracefully
			require.NoError(t, err, "type validation should not cause system error")
			require.NotNil(t, result, "result should not be nil")
		})
	}
}

func TestRejectTransactionToolErrorHandling(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	client := testEnv.GetMcpClient()
	require.NotNil(t, client, "MCP client should not be nil")

	require.NoError(t, client.Initialize(ctx), "failed to initialize MCP client")

	// Test with non-existent transaction
	args := map[string]interface{}{
		"transaction_ids": "0xnonexistent1234567890abcdef1234567890abcdef1234567890abcdef123456",
		"reason":          "suspicious_activity",
	}
	result, err := client.CallTool("reject_transaction", args)
	require.NoError(t, err, "call should not return system error")
	require.NotNil(t, result, "result should not be nil")
	require.False(t, result.IsError, "non-existent transaction should not cause tool error")

	// Validate that the result indicates failure
	markdown := getTextContent(result)
	require.True(t, strings.Contains(markdown, "Failed to reject: 1"), "should show failure count")

	// Test with mix of valid and invalid transactions
	args = map[string]interface{}{
		"transaction_ids": "0x1234567890abcdef1234567890abcdef12345678901234567890abcdef123456,0xinvalid,0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
		"reason":          "user_request",
	}
	result, err = client.CallTool("reject_transaction", args)
	require.NoError(t, err, "call should not return system error")
	require.NotNil(t, result, "result should not be nil")
	require.False(t, result.IsError, "mixed valid/invalid should not cause tool error")

	markdown = getTextContent(result)
	require.True(t, strings.Contains(markdown, "Successfully rejected") || strings.Contains(markdown, "Failed to reject"), "should show counts for mixed results")
}

func TestRejectTransactionToolPerformance(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	client := testEnv.GetMcpClient()
	require.NotNil(t, client, "MCP client should not be nil")

	require.NoError(t, client.Initialize(ctx), "failed to initialize MCP client")

	// Test response time
	start := time.Now()
	args := map[string]interface{}{
		"transaction_ids": "0x1234567890abcdef1234567890abcdef12345678901234567890abcdef123456",
		"reason":          "user_request",
	}
	result, err := client.CallTool("reject_transaction", args)
	duration := time.Since(start)

	require.NoError(t, err, "failed to call reject_transaction tool")
	require.NotNil(t, result, "reject_transaction tool result should not be nil")
	require.False(t, result.IsError, "basic request should not return error")

	// Response should be fast (under 1 second for single transaction)
	assert.Less(t, duration, 1*time.Second, "response time should be under 1 second")

	// Test with multiple transactions
	multipleIds := []string{
		"0x1234567890abcdef1234567890abcdef12345678901234567890abcdef123456",
		"0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
		"0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321",
	}

	start = time.Now()
	args = map[string]interface{}{
		"transaction_ids": strings.Join(multipleIds, ","),
		"reason":          "high_gas_fee",
		"audit_log":       true,
		"notify_user":     true,
	}
	result, err = client.CallTool("reject_transaction", args)
	duration = time.Since(start)

	require.NoError(t, err, "failed to call reject_transaction tool with multiple IDs")
	require.NotNil(t, result, "reject_transaction tool result should not be nil")
	require.False(t, result.IsError, "multiple transaction request should not return error")

	// Response should still be fast (under 2 seconds for multiple transactions with audit logging)
	assert.Less(t, duration, 2*time.Second, "response time should be under 2 seconds for multiple transactions")
}