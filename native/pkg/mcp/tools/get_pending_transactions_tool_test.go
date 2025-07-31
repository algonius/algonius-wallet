package tools

import (
	"context"
	"testing"
	"time"

	"github.com/algonius/algonius-wallet/native/pkg/wallet"
	"github.com/mark3labs/mcp-go/mcp"
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

func (m *MockWalletManager) UnlockWallet(password string) error {
	return nil
}

func (m *MockWalletManager) LockWallet() {
}

func (m *MockWalletManager) IsUnlocked() bool {
	return true
}

func (m *MockWalletManager) HasWallet() bool {
	return true
}

func (m *MockWalletManager) GetCurrentWallet() *wallet.WalletStatus {
	return &wallet.WalletStatus{}
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

func TestGetPendingTransactionsToolHandler_Success(t *testing.T) {
	// Create mock transactions
	now := time.Now()
	mockTransactions := []*wallet.PendingTransaction{
		{
			Hash:                      "0x123456789abcdef",
			Chain:                     "ethereum",
			From:                      "0x123",
			To:                        "0x456",
			Amount:                    "1.5",
			Token:                     "ETH",
			Type:                      "transfer",
			Status:                    "pending",
			Confirmations:             2,
			RequiredConfirmations:     12,
			GasFee:                    "0.002",
			Priority:                  "medium",
			EstimatedConfirmationTime: "2 minutes",
			SubmittedAt:               now,
			LastChecked:               now,
		},
	}
	
	mockManager := &MockWalletManagerWithTransactions{
		mockTransactions: mockTransactions,
	}
	
	tool := NewGetPendingTransactionsTool(mockManager)
	handler := tool.GetHandler()
	
	// Create proper request structure
	params := mcp.CallToolParams{
		Name: "get_pending_transactions",
		Arguments: map[string]interface{}{
			"chain":  "ethereum",
			"limit":  10,
			"offset": 0,
		},
	}
	
	req := mcp.CallToolRequest{}
	req.Params = params
	
	result, err := handler(context.Background(), req)
	
	// Verify the result
	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, result.IsError)
	require.NotNil(t, result.Content)
	require.Len(t, result.Content, 1)
	
	// Check that the markdown output contains expected content
	textContent, ok := mcp.AsTextContent(result.Content[0])
	require.True(t, ok, "result should contain text content")
	markdown := textContent.Text
	assert.Contains(t, markdown, "### Pending Transactions")
	assert.Contains(t, markdown, "Found 1 pending transactions")
	assert.Contains(t, markdown, "0x123456789abcdef")
}

func TestGetPendingTransactionsToolHandler_NoTransactions(t *testing.T) {
	mockManager := &MockWalletManagerWithTransactions{
		mockTransactions: []*wallet.PendingTransaction{},
	}
	
	tool := NewGetPendingTransactionsTool(mockManager)
	handler := tool.GetHandler()
	
	// Create proper request structure
	params := mcp.CallToolParams{
		Name: "get_pending_transactions",
		Arguments: map[string]interface{}{
			"chain": "ethereum",
		},
	}
	
	req := mcp.CallToolRequest{}
	req.Params = params
	
	result, err := handler(context.Background(), req)
	
	// Verify the result
	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, result.IsError)
	require.NotNil(t, result.Content)
	require.Len(t, result.Content, 1)
	
	// Check that the markdown output indicates no transactions
	textContent, ok := mcp.AsTextContent(result.Content[0])
	require.True(t, ok, "result should contain text content")
	markdown := textContent.Text
	assert.Contains(t, markdown, "### Pending Transactions")
	assert.Contains(t, markdown, "No pending transactions found")
}