package wallet

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWalletManagerSendTransactionSolanaValidAddresses(t *testing.T) {
	wm := NewWalletManager()
	wm.currentWallet = NewWalletStatus("wallet-ready", "pubkey")

	from := "FnVyf9f7hFmA6N5HtV6nQWmvMRGsiE9zraFMvx6bMpiK"
	to := "5oNDL3swdJJF1g9DzJiZ4ynHXgszjAEpUkxVYejchzrY"

	txHash, err := wm.SendTransaction(context.Background(), "solana", from, to, "0.1", "")
	require.NoError(t, err)
	require.NotEmpty(t, txHash)
}

func TestWalletManagerSendTransactionSolanaInvalidFromAddress(t *testing.T) {
	wm := NewWalletManager()
	wm.currentWallet = NewWalletStatus("wallet-ready", "pubkey")

	_, err := wm.SendTransaction(
		context.Background(),
		"solana",
		"0x1234567890123456789012345678901234567890",
		"5oNDL3swdJJF1g9DzJiZ4ynHXgszjAEpUkxVYejchzrY",
		"0.1",
		"",
	)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "invalid from address"))
}
