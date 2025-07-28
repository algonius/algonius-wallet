package tools

import (
	"context"
	"testing"

	"github.com/algonius/algonius-wallet/native/pkg/wallet"
)

// MockWalletManager is a mock implementation of wallet.IWalletManager for testing
type MockWalletManager struct{}

func (m *MockWalletManager) CreateWallet(ctx context.Context, chain string) (string, string, error) {
	return "0x1234567890123456789012345678901234567890", "0xabcdef1234567890abcdef1234567890abcdef12", nil
}

func (m *MockWalletManager) ImportWallet(ctx context.Context, mnemonic, password, chainName, derivationPath string) (string, string, int64, error) {
	return "0x1234567890123456789012345678901234567890", "0xabcdef1234567890abcdef1234567890abcdef12", 1234567890, nil
}

func (m *MockWalletManager) GetBalance(ctx context.Context, address string, token string) (string, error) {
	return "1.5", nil
}

func (m *MockWalletManager) GetStatus(ctx context.Context) (*wallet.WalletStatus, error) {
	return &wallet.WalletStatus{
		Address:   "0x1234567890123456789012345678901234567890",
		PublicKey: "0xabcdef1234567890abcdef1234567890abcdef12",
		Ready:     true,
		Chains:    make(map[string]bool),
		LastUsed:  1234567890,
	}, nil
}

func (m *MockWalletManager) SendTransaction(ctx context.Context, chain, from, to, amount, token string) (string, error) {
	return "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", nil
}

func (m *MockWalletManager) EstimateGas(ctx context.Context, chain, from, to, amount, token string) (uint64, string, error) {
	return 21000, "20", nil
}

func (m *MockWalletManager) GetPendingTransactions(ctx context.Context, chain, address, transactionType string, limit, offset int) ([]*wallet.PendingTransaction, error) {
	return []*wallet.PendingTransaction{}, nil
}

func (m *MockWalletManager) RejectTransactions(ctx context.Context, transactionIds []string, reason, details string, notifyUser, auditLog bool) ([]wallet.TransactionRejectionResult, error) {
	return []wallet.TransactionRejectionResult{}, nil
}

func (m *MockWalletManager) GetTransactionHistory(ctx context.Context, address string, fromBlock, toBlock *uint64, limit, offset int) ([]*wallet.HistoricalTransaction, error) {
	return []*wallet.HistoricalTransaction{}, nil
}

func (m *MockWalletManager) GetAccounts(ctx context.Context) ([]string, error) {
	return []string{"0x1234567890123456789012345678901234567890"}, nil
}

func (m *MockWalletManager) AddPendingTransaction(ctx context.Context, tx *wallet.PendingTransaction) error {
	return nil
}

func TestConfirmTransactionTool_GetMeta(t *testing.T) {
	mockManager := &MockWalletManager{}
	tool := NewConfirmTransactionTool(mockManager)

	meta := tool.GetMeta()

	if meta.Name != "confirm_transaction" {
		t.Errorf("Expected tool name 'confirm_transaction', got '%s'", meta.Name)
	}

	if meta.Description != "Check transaction confirmation status" {
		t.Errorf("Expected description 'Check transaction confirmation status', got '%s'", meta.Description)
	}
}

func TestConfirmTransactionTool_NormalizeChainName(t *testing.T) {
	mockManager := &MockWalletManager{}
	tool := NewConfirmTransactionTool(mockManager)

	tests := []struct {
		input    string
		expected string
		hasError bool
	}{
		{"ethereum", "ETH", false},
		{"eth", "ETH", false},
		{"ETH", "ETH", false},
		{"Ethereum", "ETH", false},
		{"bsc", "BSC", false},
		{"BSC", "BSC", false},
		{"binance", "BSC", false},
		{"unsupported", "", true},
	}

	for _, test := range tests {
		result, err := tool.normalizeChainName(test.input)
		if test.hasError {
			if err == nil {
				t.Errorf("Expected error for input '%s', but got none", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("Expected no error for input '%s', but got: %v", test.input, err)
			}
			if result != test.expected {
				t.Errorf("Expected '%s' for input '%s', got '%s'", test.expected, test.input, result)
			}
		}
	}
}