package integration

import (
	"context"
	"testing"

	"github.com/algonius/algonius-wallet/native/tests/integration/env"
	"github.com/stretchr/testify/require"
)

func TestCreateWalletTool(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	client := testEnv.GetMcpClient()
	require.NotNil(t, client, "MCP client should not be nil")

	require.NoError(t, client.Initialize(ctx), "failed to initialize MCP client")

	// Call the create_wallet tool
	args := map[string]interface{}{
		"chain": "ETH",
	}
	result, err := client.CallTool("create_wallet", args)
	require.NoError(t, err, "failed to call create_wallet tool")
	require.NotNil(t, result, "create_wallet tool result should not be nil")
}
