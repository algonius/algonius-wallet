package tools

import (
	"context"
	"testing"

	"github.com/algonius/algonius-wallet/native/pkg/wallet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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