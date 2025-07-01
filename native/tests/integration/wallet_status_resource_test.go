package integration

import (
	"context"
	"testing"

	"github.com/algonius/algonius-wallet/native/tests/integration/env"
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/stretchr/testify/require"
)

func TestWalletStatusResource(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	client := testEnv.GetMcpClient()
	require.NotNil(t, client, "MCP client should not be nil")

	require.NoError(t, client.Initialize(ctx), "failed to initialize MCP client")

	result, err := client.ReadResource("wallet://status")
	require.NoError(t, err, "failed to read wallet://status resource")
	require.NotNil(t, result, "wallet://status resource result should not be nil")

	// Verify the result has content
	require.NotEmpty(t, result.Contents, "wallet status resource should have content")

	// Type assert to TextResourceContents to access the fields
	content, ok := result.Contents[0].(mcp.TextResourceContents)
	require.True(t, ok, "content should be TextResourceContents type")
	require.Equal(t, "text/markdown", content.MIMEType, "content should be Markdown")

	// Verify the Markdown content structure
	markdownText := content.Text
	require.Contains(t, markdownText, "# Wallet Status", "should contain wallet status header")
	require.Contains(t, markdownText, "## Overview", "should contain overview section")
	require.Contains(t, markdownText, "## Supported Chains", "should contain supported chains section")

	// For a fresh wallet manager with no wallet created yet, ready should be "Not Ready"
	require.Contains(t, markdownText, "**Status**: Not Ready", "wallet should not be ready initially")
	require.Contains(t, markdownText, "**Address**: Not created yet", "address should show not created initially")
	require.Contains(t, markdownText, "**Public Key**: Not created yet", "public key should show not created initially")
	require.Contains(t, markdownText, "**Last Used**: Never", "last used should show never initially")

	// Validate that supported chains are present with checkmarks
	require.Contains(t, markdownText, "✅ Ethereum (ETH)", "should show Ethereum as supported")
	require.Contains(t, markdownText, "✅ Binance Smart Chain (BSC)", "should show BSC as supported")
	require.Contains(t, markdownText, "✅ Solana (SOL)", "should show Solana as supported")

	t.Logf("wallet://status resource result: %+v", result)
	t.Logf("wallet status markdown:\n%s", content.Text)
}

func TestWalletStatusResourceAfterWalletCreation(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	client := testEnv.GetMcpClient()
	require.NotNil(t, client, "MCP client should not be nil")

	require.NoError(t, client.Initialize(ctx), "failed to initialize MCP client")

	// First create a wallet
	args := map[string]interface{}{
		"chain": "ETH",
	}
	walletResult, err := client.CallTool("create_wallet", args)
	require.NoError(t, err, "failed to call create_wallet tool")
	require.NotNil(t, walletResult, "create_wallet tool result should not be nil")

	// Now check the wallet status
	result, err := client.ReadResource("wallet://status")
	require.NoError(t, err, "failed to read wallet://status resource")
	require.NotNil(t, result, "wallet://status resource result should not be nil")

	// Parse the wallet status from Markdown
	content, ok := result.Contents[0].(mcp.TextResourceContents)
	require.True(t, ok, "content should be TextResourceContents type")
	require.Equal(t, "text/markdown", content.MIMEType, "content should be Markdown")

	markdownText := content.Text

	// After wallet creation, it should be ready and have address/public key
	require.Contains(t, markdownText, "**Status**: Ready", "wallet should be ready after creation")
	require.NotContains(t, markdownText, "**Address**: Not created yet", "address should be set after creation")
	require.NotContains(t, markdownText, "**Public Key**: Not created yet", "public key should be set after creation")
	require.NotContains(t, markdownText, "**Last Used**: Never", "last used should be set after creation")

	// Check that it contains an actual address (starts with 0x for Ethereum)
	require.Contains(t, markdownText, "**Address**: 0x", "should contain actual address starting with 0x")

	t.Logf("wallet status after creation:\n%s", markdownText)
}