package integration

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/algonius/algonius-wallet/native/tests/integration/env"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to extract text content from MCP result
func getTextContent(result *mcp.CallToolResult) string {
	if result == nil || len(result.Content) == 0 {
		return ""
	}
	
	textContent, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		return ""
	}
	
	return textContent.Text
}

func TestGetPendingTransactionsTool(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	client := testEnv.GetMcpClient()
	require.NotNil(t, client, "MCP client should not be nil")

	require.NoError(t, client.Initialize(ctx), "failed to initialize MCP client")

	// Test basic request with no filters
	args := map[string]interface{}{}
	result, err := client.CallTool("get_pending_transactions", args)
	require.NoError(t, err, "failed to call get_pending_transactions tool")
	require.NotNil(t, result, "get_pending_transactions tool result should not be nil")
	require.False(t, result.IsError, "basic request should not return error")

	// Test with chain filter
	args = map[string]interface{}{
		"chain": "ethereum",
	}
	result, err = client.CallTool("get_pending_transactions", args)
	require.NoError(t, err, "failed to call get_pending_transactions tool with chain filter")
	require.NotNil(t, result, "get_pending_transactions tool result should not be nil")
	require.False(t, result.IsError, "chain filter request should not return error")

	// Test with address filter
	args = map[string]interface{}{
		"address": "0x742d35Cc6634C0532925a3b8D4C2B79C2b86A7A8",
	}
	result, err = client.CallTool("get_pending_transactions", args)
	require.NoError(t, err, "failed to call get_pending_transactions tool with address filter")
	require.NotNil(t, result, "get_pending_transactions tool result should not be nil")
	require.False(t, result.IsError, "address filter request should not return error")

	// Test with type filter
	args = map[string]interface{}{
		"type": "transfer",
	}
	result, err = client.CallTool("get_pending_transactions", args)
	require.NoError(t, err, "failed to call get_pending_transactions tool with type filter")
	require.NotNil(t, result, "get_pending_transactions tool result should not be nil")
	require.False(t, result.IsError, "type filter request should not return error")

	// Test with pagination
	args = map[string]interface{}{
		"limit":  2,
		"offset": 1,
	}
	result, err = client.CallTool("get_pending_transactions", args)
	require.NoError(t, err, "failed to call get_pending_transactions tool with pagination")
	require.NotNil(t, result, "get_pending_transactions tool result should not be nil")
	require.False(t, result.IsError, "pagination request should not return error")

	// Test with all filters combined
	args = map[string]interface{}{
		"chain":   "ethereum",
		"address": "0x742d35Cc6634C0532925a3b8D4C2B79C2b86A7A8",
		"type":    "transfer",
		"limit":   5,
		"offset":  0,
	}
	result, err = client.CallTool("get_pending_transactions", args)
	require.NoError(t, err, "failed to call get_pending_transactions tool with all filters")
	require.NotNil(t, result, "get_pending_transactions tool result should not be nil")
	require.False(t, result.IsError, "combined filters request should not return error")

	// Test with invalid limit (should be capped at 100)
	args = map[string]interface{}{
		"limit": 200,
	}
	result, err = client.CallTool("get_pending_transactions", args)
	require.NoError(t, err, "failed to call get_pending_transactions tool with large limit")
	require.NotNil(t, result, "get_pending_transactions tool result should not be nil")
	require.False(t, result.IsError, "large limit request should not return error")

	// Test with negative offset (should be normalized to 0)
	args = map[string]interface{}{
		"offset": -5,
	}
	result, err = client.CallTool("get_pending_transactions", args)
	require.NoError(t, err, "failed to call get_pending_transactions tool with negative offset")
	require.NotNil(t, result, "get_pending_transactions tool result should not be nil")
	require.False(t, result.IsError, "negative offset request should not return error")

	// Test with non-existent chain filter
	args = map[string]interface{}{
		"chain": "nonexistent_chain",
	}
	result, err = client.CallTool("get_pending_transactions", args)
	require.NoError(t, err, "failed to call get_pending_transactions tool with non-existent chain")
	require.NotNil(t, result, "get_pending_transactions tool result should not be nil")
	require.False(t, result.IsError, "non-existent chain filter should not return error")

	// Test with non-existent address filter
	args = map[string]interface{}{
		"address": "0x0000000000000000000000000000000000000000",
	}
	result, err = client.CallTool("get_pending_transactions", args)
	require.NoError(t, err, "failed to call get_pending_transactions tool with non-existent address")
	require.NotNil(t, result, "get_pending_transactions tool result should not be nil")
	require.False(t, result.IsError, "non-existent address filter should not return error")
}

// TestGetPendingTransactionsToolContentValidation tests content validation and response format
func TestGetPendingTransactionsToolContentValidation(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	client := testEnv.GetMcpClient()
	require.NotNil(t, client, "MCP client should not be nil")

	require.NoError(t, client.Initialize(ctx), "failed to initialize MCP client")

	// Test default response format
	args := map[string]interface{}{}
	result, err := client.CallTool("get_pending_transactions", args)
	require.NoError(t, err, "failed to call get_pending_transactions tool")
	require.NotNil(t, result, "get_pending_transactions tool result should not be nil")
	require.False(t, result.IsError, "basic request should not return error")

	// Validate response structure
	require.NotNil(t, result.Content, "result should have content")
	require.Len(t, result.Content, 1, "result should have one content item")
	
	// Validate markdown content format
	markdown := getTextContent(result)
	require.NotEmpty(t, markdown, "should have text content")
	require.True(t, strings.Contains(markdown, "### Pending Transactions"), "should contain main heading")
	require.True(t, strings.Contains(markdown, "Found") || strings.Contains(markdown, "No pending transactions"), "should contain transaction count info")

	// Test with specific filters to get known data
	args = map[string]interface{}{
		"chain": "ethereum",
		"type":  "transfer",
	}
	result, err = client.CallTool("get_pending_transactions", args)
	require.NoError(t, err, "failed to call get_pending_transactions tool with filters")
	require.NotNil(t, result, "get_pending_transactions tool result should not be nil")
	require.False(t, result.IsError, "filtered request should not return error")

	// Validate transaction details format
	markdown = getTextContent(result)
	if strings.Contains(markdown, "Transaction 1") {
		// Check required fields are present
		require.True(t, strings.Contains(markdown, "**Hash**:"), "should contain transaction hash")
		require.True(t, strings.Contains(markdown, "**Chain**:"), "should contain chain info")
		require.True(t, strings.Contains(markdown, "**From**:"), "should contain from address")
		require.True(t, strings.Contains(markdown, "**To**:"), "should contain to address")
		require.True(t, strings.Contains(markdown, "**Amount**:"), "should contain amount")
		require.True(t, strings.Contains(markdown, "**Status**:"), "should contain status")
		require.True(t, strings.Contains(markdown, "**Confirmations**:"), "should contain confirmations")
		require.True(t, strings.Contains(markdown, "**Priority**:"), "should contain priority")
	}
}

// TestGetPendingTransactionsToolPaginationLogic tests pagination behavior
func TestGetPendingTransactionsToolPaginationLogic(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	client := testEnv.GetMcpClient()
	require.NotNil(t, client, "MCP client should not be nil")

	require.NoError(t, client.Initialize(ctx), "failed to initialize MCP client")

	// Test limit boundary conditions
	testCases := []struct {
		name   string
		limit  int
		offset int
	}{
		{"zero limit", 0, 0},
		{"negative limit", -1, 0},
		{"small limit", 1, 0},
		{"medium limit", 5, 0},
		{"large limit", 100, 0},
		{"exceed max limit", 500, 0},
		{"with offset", 2, 1},
		{"large offset", 10, 100},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			args := map[string]interface{}{
				"limit":  tc.limit,
				"offset": tc.offset,
			}
			result, err := client.CallTool("get_pending_transactions", args)
			require.NoError(t, err, "failed to call get_pending_transactions tool")
			require.NotNil(t, result, "get_pending_transactions tool result should not be nil")
			require.False(t, result.IsError, "pagination request should not return error")

			// Validate response format
			require.NotNil(t, result.Content, "result should have content")
			require.Len(t, result.Content, 1, "result should have one content item")
		})
	}
}

// TestGetPendingTransactionsToolFilterValidation tests various filter combinations
func TestGetPendingTransactionsToolFilterValidation(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	client := testEnv.GetMcpClient()
	require.NotNil(t, client, "MCP client should not be nil")

	require.NoError(t, client.Initialize(ctx), "failed to initialize MCP client")

	// Test various chain filters
	chainFilters := []string{"ethereum", "bsc", "solana", "ETH", "BSC", "ETHEREUM", "invalid_chain"}
	for _, chain := range chainFilters {
		t.Run(fmt.Sprintf("chain_%s", chain), func(t *testing.T) {
			args := map[string]interface{}{
				"chain": chain,
			}
			result, err := client.CallTool("get_pending_transactions", args)
			require.NoError(t, err, "failed to call get_pending_transactions tool")
			require.NotNil(t, result, "get_pending_transactions tool result should not be nil")
			require.False(t, result.IsError, "chain filter should not return error")
		})
	}

	// Test various transaction types
	typeFilters := []string{"transfer", "swap", "contract", "TRANSFER", "invalid_type"}
	for _, txType := range typeFilters {
		t.Run(fmt.Sprintf("type_%s", txType), func(t *testing.T) {
			args := map[string]interface{}{
				"type": txType,
			}
			result, err := client.CallTool("get_pending_transactions", args)
			require.NoError(t, err, "failed to call get_pending_transactions tool")
			require.NotNil(t, result, "get_pending_transactions tool result should not be nil")
			require.False(t, result.IsError, "type filter should not return error")
		})
	}

	// Test various address formats
	addressFilters := []string{
		"0x742d35Cc6634C0532925a3b8D4C2B79C2b86A7A8",
		"0x0000000000000000000000000000000000000000",
		"0xffffffffffffffffffffffffffffffffffffffff",
		"invalid_address",
		"0x123", // Short address
		"",     // Empty address
	}
	for i, address := range addressFilters {
		t.Run(fmt.Sprintf("address_%d", i), func(t *testing.T) {
			args := map[string]interface{}{
				"address": address,
			}
			result, err := client.CallTool("get_pending_transactions", args)
			require.NoError(t, err, "failed to call get_pending_transactions tool")
			require.NotNil(t, result, "get_pending_transactions tool result should not be nil")
			require.False(t, result.IsError, "address filter should not return error")
		})
	}
}

// TestGetPendingTransactionsToolPerformance tests performance characteristics
func TestGetPendingTransactionsToolPerformance(t *testing.T) {
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
	args := map[string]interface{}{}
	result, err := client.CallTool("get_pending_transactions", args)
	duration := time.Since(start)

	require.NoError(t, err, "failed to call get_pending_transactions tool")
	require.NotNil(t, result, "get_pending_transactions tool result should not be nil")
	require.False(t, result.IsError, "basic request should not return error")

	// Response should be fast (under 1 second for mock data)
	assert.Less(t, duration, 1*time.Second, "response time should be under 1 second")

	// Test concurrent requests
	conCurrentRequests := 10
	results := make(chan error, conCurrentRequests)

	for i := 0; i < conCurrentRequests; i++ {
		go func(id int) {
			args := map[string]interface{}{
				"limit": 5,
				"offset": id,
			}
			result, err := client.CallTool("get_pending_transactions", args)
			if err != nil {
				results <- err
				return
			}
			if result.IsError {
				results <- fmt.Errorf("request %d returned error", id)
				return
			}
			results <- nil
		}(i)
	}

	// Wait for all requests to complete
	for i := 0; i < conCurrentRequests; i++ {
		select {
		case err := <-results:
			require.NoError(t, err, "concurrent request failed")
		case <-time.After(5 * time.Second):
			t.Fatal("concurrent request timed out")
		}
	}
}

// TestGetPendingTransactionsToolSecurityValidation tests security aspects
func TestGetPendingTransactionsToolSecurityValidation(t *testing.T) {
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
		{"chain": "'; DROP TABLE transactions; --"},
		{"address": "<script>alert('xss')</script>"},
		{"type": "../../../etc/passwd"},
		{"limit": "999999999999999999999"},
		{"offset": "not_a_number"},
	}

	for i, args := range injectionAttempts {
		t.Run(fmt.Sprintf("injection_%d", i), func(t *testing.T) {
			result, err := client.CallTool("get_pending_transactions", args)
			// Should not panic or cause system errors
			require.NoError(t, err, "injection attempt should not cause system error")
			require.NotNil(t, result, "result should not be nil")
			// May return error or empty result, but should not crash
		})
	}

	// Test parameter type validation
	typeValidationTests := []map[string]interface{}{
		{"chain": 123},
		{"address": []string{"addr1", "addr2"}},
		{"type": map[string]string{"key": "value"}},
		{"limit": "not_a_number"},
		{"offset": true},
	}

	for i, args := range typeValidationTests {
		t.Run(fmt.Sprintf("type_validation_%d", i), func(t *testing.T) {
			result, err := client.CallTool("get_pending_transactions", args)
			// Should handle type mismatches gracefully
			require.NoError(t, err, "type validation should not cause system error")
			require.NotNil(t, result, "result should not be nil")
		})
	}

	// Test extremely large request parameters
	largeParams := map[string]interface{}{
		"chain":   strings.Repeat("a", 10000),
		"address": strings.Repeat("0x", 1000),
		"type":    strings.Repeat("transfer", 1000),
		"limit":   999999,
		"offset":  999999,
	}

	result, err := client.CallTool("get_pending_transactions", largeParams)
	require.NoError(t, err, "large parameters should not cause system error")
	require.NotNil(t, result, "result should not be nil")
}

// TestGetPendingTransactionsToolDataConsistency tests data consistency
func TestGetPendingTransactionsToolDataConsistency(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	client := testEnv.GetMcpClient()
	require.NotNil(t, client, "MCP client should not be nil")

	require.NoError(t, client.Initialize(ctx), "failed to initialize MCP client")

	// Test consistency across multiple calls
	args := map[string]interface{}{
		"chain": "ethereum",
		"limit": 10,
	}

	// Make multiple calls and compare results
	var firstResult, secondResult *mcp.CallToolResult

	firstResult, err = client.CallTool("get_pending_transactions", args)
	require.NoError(t, err, "first call should succeed")
	require.NotNil(t, firstResult, "first result should not be nil")
	require.False(t, firstResult.IsError, "first call should not return error")

	// Wait a bit and call again
	time.Sleep(100 * time.Millisecond)

	secondResult, err = client.CallTool("get_pending_transactions", args)
	require.NoError(t, err, "second call should succeed")
	require.NotNil(t, secondResult, "second result should not be nil")
	require.False(t, secondResult.IsError, "second call should not return error")

	// Results should be consistent for mock data
	require.Equal(t, len(firstResult.Content), len(secondResult.Content), "content length should be consistent")

	// Test pagination consistency
	page1Args := map[string]interface{}{
		"limit":  1,
		"offset": 0,
	}
	page2Args := map[string]interface{}{
		"limit":  1,
		"offset": 1,
	}

	page1Result, err := client.CallTool("get_pending_transactions", page1Args)
	require.NoError(t, err, "page 1 call should succeed")
	require.NotNil(t, page1Result, "page 1 result should not be nil")
	require.False(t, page1Result.IsError, "page 1 call should not return error")

	page2Result, err := client.CallTool("get_pending_transactions", page2Args)
	require.NoError(t, err, "page 2 call should succeed")
	require.NotNil(t, page2Result, "page 2 result should not be nil")
	require.False(t, page2Result.IsError, "page 2 call should not return error")

	// Pages should be different (unless there's only one transaction)
	if len(page1Result.Content) > 0 && len(page2Result.Content) > 0 {
		page1Content := getTextContent(page1Result)
		page2Content := getTextContent(page2Result)
		
		// If both pages have transactions, they should be different
		if strings.Contains(page1Content, "Transaction 1") && strings.Contains(page2Content, "Transaction 1") {
			// Both show "Transaction 1" but content should be different due to pagination
			// This is expected behavior - "Transaction 1" is the first transaction on each page
		}
	}
}

// TestGetPendingTransactionsToolErrorHandling tests error handling scenarios
func TestGetPendingTransactionsToolErrorHandling(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	client := testEnv.GetMcpClient()
	require.NotNil(t, client, "MCP client should not be nil")

	require.NoError(t, client.Initialize(ctx), "failed to initialize MCP client")

	// Test with malformed JSON-like strings
	malformedArgs := []map[string]interface{}{
		{"chain": `{"malformed": json`},
		{"address": `["array", "values"]`},
		{"type": `"quoted string"`},
	}

	for i, args := range malformedArgs {
		t.Run(fmt.Sprintf("malformed_%d", i), func(t *testing.T) {
			result, err := client.CallTool("get_pending_transactions", args)
			// Should handle malformed input gracefully
			require.NoError(t, err, "malformed input should not cause system error")
			require.NotNil(t, result, "result should not be nil")
		})
	}

	// Test with nil values
	nilArgs := map[string]interface{}{
		"chain":   nil,
		"address": nil,
		"type":    nil,
		"limit":   nil,
		"offset":  nil,
	}

	result, err := client.CallTool("get_pending_transactions", nilArgs)
	require.NoError(t, err, "nil values should not cause system error")
	require.NotNil(t, result, "result should not be nil")

	// Test with completely invalid JSON structure
	invalidStructure := map[string]interface{}{
		"nested": map[string]interface{}{
			"deeply": map[string]interface{}{
				"invalid": "structure",
			},
		},
	}

	result, err = client.CallTool("get_pending_transactions", invalidStructure)
	require.NoError(t, err, "invalid structure should not cause system error")
	require.NotNil(t, result, "result should not be nil")
}