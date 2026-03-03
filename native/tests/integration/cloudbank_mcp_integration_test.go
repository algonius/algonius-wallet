//go:build integration
// +build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/algonius/algonius-wallet/native/tests/integration/env"
	"github.com/stretchr/testify/require"
)

func TestCloudBankMCPIntegration(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(&env.TestConfig{
		MockMode: true,
	})
	require.NoError(t, err, "failed to create CloudBank integration test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup CloudBank integration test environment")

	client := testEnv.GetMcpClient()
	require.NotNil(t, client, "MCP client should not be nil")
	require.NoError(t, client.Initialize(ctx), "failed to initialize MCP client")

	fromAddress := ""
	txHash := ""

	t.Run("create_wallet", func(t *testing.T) {
		result := mustCallToolSuccess(t, client, "create_wallet", map[string]interface{}{
			"chain": "bsc",
		})

		text := getTextContent(result)
		require.Contains(t, text, "### Wallet Created")

		var extractErr error
		fromAddress, extractErr = extractAddress(text)
		require.NoError(t, extractErr)
		require.NotEmpty(t, fromAddress)
	})

	t.Run("fund_via_cloudbank_testnet_faucet", func(t *testing.T) {
		fundingRef, fundingErr := requestCloudBankFaucetFunding(ctx, fromAddress)
		require.NoError(t, fundingErr)
		require.NotEmpty(t, fundingRef)
	})

	t.Run("get_balance", func(t *testing.T) {
		result := mustCallToolSuccess(t, client, "get_balance", map[string]interface{}{
			"address": fromAddress,
			"token":   "BNB",
		})
		require.Contains(t, getTextContent(result), "### Wallet Balance")
	})

	t.Run("sign_transaction_payload", func(t *testing.T) {
		result := mustCallToolSuccess(t, client, "sign_message", map[string]interface{}{
			"address": fromAddress,
			"message": fmt.Sprintf("cloudbank-predict-order:%s:%d", strings.ToLower(fromAddress), time.Now().Unix()),
		})
		require.Contains(t, getTextContent(result), "### Message Signed Successfully")
	})

	t.Run("send_transaction", func(t *testing.T) {
		result := mustCallToolSuccess(t, client, "send_transaction", map[string]interface{}{
			"chain":  "bsc",
			"from":   fromAddress,
			"to":     "0x907f4DAA6Ff8083EBdb60FC548603bA79DC970f6", // CloudBank test tUSDT address (BSC testnet)
			"amount": "0.001",
			"token":  "BNB",
		})
		text := getTextContent(result)
		require.Contains(t, text, "### Transaction Sent")
		require.Contains(t, text, "- **Chain**: `bsc`")

		var extractErr error
		txHash, extractErr = extractTransactionHash(text)
		require.NoError(t, extractErr)
		require.NotEmpty(t, txHash)
	})

	t.Run("confirm_transaction", func(t *testing.T) {
		result := mustCallToolSuccess(t, client, "get_transaction_status", map[string]interface{}{
			"transaction_hash": txHash,
			"chain":            "bsc",
		})
		require.Contains(t, getTextContent(result), "### Transaction Status")
	})

	t.Run("stability_consecutive_calls", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			msgResult := mustCallToolSuccess(t, client, "sign_message", map[string]interface{}{
				"address": fromAddress,
				"message": fmt.Sprintf("cloudbank-stability-%d", i),
			})
			require.Contains(t, getTextContent(msgResult), "### Message Signed Successfully")
		}

		for i := 0; i < 5; i++ {
			balanceResult := mustCallToolSuccess(t, client, "get_balance", map[string]interface{}{
				"address": fromAddress,
				"token":   "BNB",
			})
			require.Contains(t, getTextContent(balanceResult), "### Wallet Balance")
		}
	})

	t.Run("error_recovery", func(t *testing.T) {
		errResult := mustCallToolResult(t, client, "send_transaction", map[string]interface{}{
			"chain":  "bsc",
			"from":   "invalid-cloudbank-address",
			"to":     "0x907f4DAA6Ff8083EBdb60FC548603bA79DC970f6",
			"amount": "0.001",
			"token":  "BNB",
		})
		require.True(t, errResult.IsError, "invalid address should return tool error")
		require.Contains(t, strings.ToLower(getTextContent(errResult)), "invalid from address")

		recoveryResult := mustCallToolSuccess(t, client, "send_transaction", map[string]interface{}{
			"chain":  "bsc",
			"from":   fromAddress,
			"to":     "0x6299960264AC6c64592AcAaad96b647d0BaeF1C1", // CloudBank test tCOD address (BSC testnet)
			"amount": "0.001",
			"token":  "BNB",
		})
		require.Contains(t, getTextContent(recoveryResult), "### Transaction Sent")
	})
}

func requestCloudBankFaucetFunding(ctx context.Context, address string) (string, error) {
	endpoint := strings.TrimSpace(os.Getenv("CLOUDBANK_FAUCET_API_URL"))
	if endpoint == "" {
		return "mock-faucet-funding", nil
	}

	requestPayload := map[string]string{
		"address": address,
		"chain":   "bsc-testnet",
		"source":  "algonius-wallet-kr4",
	}

	payloadBytes, err := json.Marshal(requestPayload)
	if err != nil {
		return "", fmt.Errorf("marshal faucet request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("build faucet request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("call faucet endpoint: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("faucet request failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	if len(body) == 0 {
		return "faucet-request-submitted", nil
	}

	var responseMap map[string]interface{}
	if err := json.Unmarshal(body, &responseMap); err != nil {
		return "faucet-request-submitted", nil
	}

	for _, key := range []string{"txHash", "tx_hash", "hash"} {
		value, ok := responseMap[key]
		if !ok {
			continue
		}
		hash, ok := value.(string)
		if ok && strings.TrimSpace(hash) != "" {
			return hash, nil
		}
	}

	return "faucet-request-submitted", nil
}
