//go:build integration
// +build integration

package integration

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/algonius/algonius-wallet/native/tests/integration/env"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

type transactionFlowCase struct {
	chainLabel  string
	createChain string
	txChain     string
	to          string
	token       string
	invalidFrom string
}

func TestE2ETransactionFlowETH(t *testing.T) {
	runTransactionFlowE2E(t, transactionFlowCase{
		chainLabel:  "ETH",
		createChain: "ethereum",
		txChain:     "eth",
		to:          "0x8ba1f109551bd432803012645ac136ddd64dba72",
		token:       "ETH",
		invalidFrom: "invalid-eth-address",
	})
}

func TestE2ETransactionFlowBSC(t *testing.T) {
	runTransactionFlowE2E(t, transactionFlowCase{
		chainLabel:  "BSC",
		createChain: "bsc",
		txChain:     "binance",
		to:          "0x0987654321098765432109876543210987654321",
		token:       "BNB",
		invalidFrom: "invalid-bsc-address",
	})
}

func TestE2ETransactionFlowSOL(t *testing.T) {
	runTransactionFlowE2E(t, transactionFlowCase{
		chainLabel:  "SOL",
		createChain: "solana",
		txChain:     "sol",
		to:          "5oNDL3swdJJF1g9DzJiZ4ynHXgszjAEpUkxVYejchzrY",
		token:       "SOL",
		invalidFrom: "0x1234567890123456789012345678901234567890",
	})
}

func runTransactionFlowE2E(t *testing.T, tc transactionFlowCase) {
	t.Helper()

	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(&env.TestConfig{
		MockMode: true,
	})
	require.NoError(t, err, "failed to create test environment for %s", tc.chainLabel)
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment for %s", tc.chainLabel)

	client := testEnv.GetMcpClient()
	require.NotNil(t, client, "MCP client should not be nil")
	require.NoError(t, client.Initialize(ctx), "failed to initialize MCP client")

	fromAddress := ""
	t.Run("create_wallet", func(t *testing.T) {
		result := mustCallToolSuccess(t, client, "create_wallet", map[string]interface{}{
			"chain": tc.createChain,
		})
		text := getTextContent(result)
		require.Contains(t, text, "### Wallet Created")
		var extractErr error
		fromAddress, extractErr = extractAddress(text)
		require.NoError(t, extractErr)
		require.NotEmpty(t, fromAddress)
	})

	t.Run("get_balance", func(t *testing.T) {
		result := mustCallToolSuccess(t, client, "get_balance", map[string]interface{}{
			"address": fromAddress,
			"token":   tc.token,
		})
		text := getTextContent(result)
		require.Contains(t, text, "### Wallet Balance")
		require.Contains(t, text, tc.token)
	})

	t.Run("sign_message", func(t *testing.T) {
		result := mustCallToolSuccess(t, client, "sign_message", map[string]interface{}{
			"address": fromAddress,
			"message": fmt.Sprintf("e2e-%s-sign-message", strings.ToLower(tc.chainLabel)),
		})
		require.Contains(t, getTextContent(result), "### Message Signed Successfully")
	})

	txHash := ""
	t.Run("send_transaction", func(t *testing.T) {
		result := mustCallToolSuccess(t, client, "send_transaction", map[string]interface{}{
			"chain":  tc.txChain,
			"from":   fromAddress,
			"to":     tc.to,
			"amount": "0.1",
			"token":  tc.token,
		})
		text := getTextContent(result)
		require.Contains(t, text, "### Transaction Sent")
		require.Contains(t, text, "- **Chain**: `"+normalizeForAssertion(tc.txChain)+"`")

		var extractErr error
		txHash, extractErr = extractTransactionHash(text)
		require.NoError(t, extractErr)
		require.NotEmpty(t, txHash)
	})

	t.Run("get_transaction_status", func(t *testing.T) {
		result := mustCallToolSuccess(t, client, "get_transaction_status", map[string]interface{}{
			"transaction_hash": txHash,
			"chain":            tc.txChain,
		})
		require.Contains(t, getTextContent(result), "### Transaction Status")
	})

	t.Run("get_transaction_history", func(t *testing.T) {
		result := mustCallToolSuccess(t, client, "get_transaction_history", map[string]interface{}{
			"address": fromAddress,
			"limit":   10,
		})
		require.Contains(t, getTextContent(result), "### Transaction History")
	})

	t.Run("error_path_invalid_from_address", func(t *testing.T) {
		result := mustCallToolResult(t, client, "send_transaction", map[string]interface{}{
			"chain":  tc.txChain,
			"from":   tc.invalidFrom,
			"to":     tc.to,
			"amount": "0.1",
			"token":  tc.token,
		})
		require.True(t, result.IsError, "invalid address path should return tool error")
		require.Contains(t, strings.ToLower(getTextContent(result)), "invalid from address")
	})
}

func mustCallToolSuccess(t *testing.T, client *env.McpHTTPClient, tool string, args map[string]interface{}) *mcp.CallToolResult {
	t.Helper()
	result := mustCallToolResult(t, client, tool, args)
	require.False(t, result.IsError, "tool %s should succeed", tool)
	return result
}

func mustCallToolResult(t *testing.T, client *env.McpHTTPClient, tool string, args map[string]interface{}) *mcp.CallToolResult {
	t.Helper()
	result, err := client.CallTool(tool, args)
	require.NoError(t, err, "failed to call tool %s", tool)
	require.NotNil(t, result, "tool result should not be nil")
	return result
}

func extractTransactionHash(markdown string) (string, error) {
	re := regexp.MustCompile(`\*\*Transaction Hash\*\*:\s*` + "`" + `([^` + "`" + `]+)` + "`")
	matches := re.FindStringSubmatch(markdown)
	if len(matches) < 2 {
		return "", fmt.Errorf("transaction hash not found in tool output")
	}
	return matches[1], nil
}

func extractAddress(markdown string) (string, error) {
	re := regexp.MustCompile(`\*\*Address\*\*:\s*` + "`" + `([^` + "`" + `]+)` + "`")
	matches := re.FindStringSubmatch(markdown)
	if len(matches) < 2 {
		return "", fmt.Errorf("address not found in tool output")
	}
	return matches[1], nil
}

func normalizeForAssertion(chain string) string {
	switch strings.ToLower(strings.TrimSpace(chain)) {
	case "eth", "ethereum":
		return "ethereum"
	case "bsc", "binance", "binance smart chain":
		return "bsc"
	case "sol", "solana":
		return "solana"
	default:
		return strings.ToLower(strings.TrimSpace(chain))
	}
}
