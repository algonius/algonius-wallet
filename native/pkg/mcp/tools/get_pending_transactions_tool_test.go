package tools

import (
	"context"
	"testing"

	"github.com/algonius/algonius-wallet/native/pkg/wallet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockWalletManager is a mock implementation of wallet.IWalletManager for testing
type MockWalletManager struct{}

func (m *MockWalletManager) CreateWallet(ctx context.Context, chain string) (address string, publicKey string, err error) {
	return "0x123", "0x456", nil
}

func (m *MockWalletManager) ImportWallet(ctx context.Context, mnemonic, password, chainName, derivationPath string) (address string, publicKey string, importedAt int64, err error) {
	return "0x123", "0x456", 1234567890, nil
}

func (m *MockWalletManager) GetBalance(ctx context.Context, address, token string) (string, error) {
	return "0", nil
}

func (m *MockWalletManager) GetStatus(ctx context.Context) (*wallet.WalletStatus, error) {
	return nil, nil
}

func (m *MockWalletManager) SendTransaction(ctx context.Context, chain, from, to, amount, token string) (string, error) {
	return "0xmockhash", nil
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
	return []string{}, nil
}

func (m *MockWalletManager) AddPendingTransaction(ctx context.Context, tx *wallet.PendingTransaction) error {
	return nil
}

// MockWalletManagerWithTransactions extends MockWalletManager to include mock transactions
type MockWalletManagerWithTransactions struct {
	MockWalletManager
	mockTransactions []*wallet.PendingTransaction
}

func (m *MockWalletManagerWithTransactions) GetPendingTransactions(ctx context.Context, chain, address, transactionType string, limit, offset int) ([]*wallet.PendingTransaction, error) {
	return m.mockTransactions, nil
}

func (m *MockWalletManagerWithTransactions) AddPendingTransaction(ctx context.Context, tx *wallet.PendingTransaction) error {
	return nil
}

func TestGetPendingTransactionsToolMeta(t *testing.T) {
	mockManager := &MockWalletManager{}
	tool := NewGetPendingTransactionsTool(mockManager)
	
	meta := tool.GetMeta()
	
	assert.Equal(t, "get_pending_transactions", meta.Name)
	assert.Equal(t, "Query pending transactions that require confirmation", meta.Description)
	
	// Check that all expected parameters are present
	assert.Contains(t, meta.InputSchema.Properties, "chain")
	assert.Contains(t, meta.InputSchema.Properties, "address")
	assert.Contains(t, meta.InputSchema.Properties, "type")
	assert.Contains(t, meta.InputSchema.Properties, "limit")
	assert.Contains(t, meta.InputSchema.Properties, "offset")
}

func TestGetPendingTransactionsToolCreation(t *testing.T) {
	mockManager := &MockWalletManager{}
	tool := NewGetPendingTransactionsTool(mockManager)
	
	require.NotNil(t, tool)
	require.NotNil(t, tool.manager)
	
	// Test that the tool has the expected methods
	meta := tool.GetMeta()
	require.NotNil(t, meta)
	
	handler := tool.GetHandler()
	require.NotNil(t, handler)
}