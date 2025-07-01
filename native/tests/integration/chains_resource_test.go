package integration

import (
	"context"
	"testing"

	"github.com/algonius/algonius-wallet/native/tests/integration/env"

	"github.com/stretchr/testify/require"
)

func TestChainsSupportedResource(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	client := testEnv.GetMcpClient()
	require.NotNil(t, client, "MCP client should not be nil")

	require.NoError(t, client.Initialize(ctx), "failed to initialize MCP client")

	result, err := client.ReadResource("chains://supported")
	require.NoError(t, err, "failed to read chains://supported resource")
	require.NotNil(t, result, "chains://supported resource result should not be nil")

	// Optionally, print or assert on result fields if schema is known
	t.Logf("chains://supported resource result: %+v", result)
}
